package pvz

import (
	"context"
	"strings"
	"time"

	"pvz-service/internal/domain/pvz"
	"pvz-service/internal/usecase/ports"

	"github.com/google/uuid"
)

type Service struct {
	pvzRepo       ports.PVZRepository
	receptionRepo ports.ReceptionRepository
	productRepo   ports.ProductRepository
	clock         ports.Clock
}

func NewService(pvzRepo ports.PVZRepository, receptionRepo ports.ReceptionRepository, productRepo ports.ProductRepository, clock ports.Clock) *Service {
	return &Service{pvzRepo: pvzRepo, receptionRepo: receptionRepo, productRepo: productRepo, clock: clock}
}

func (s *Service) Create(ctx context.Context, city string) (*PVZInfo, error) {
	normalizedCity, ok := normalizeCity(city)
	if !ok {
		return nil, pvz.ErrCityNotAllowed
	}
	id := uuid.New().String()
	p := &pvz.PVZ{
		ID:        id,
		City:      normalizedCity,
		CreatedAt: s.clock.Now(),
	}
	if err := s.pvzRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return &PVZInfo{
		ID:         p.ID,
		City:       p.City,
		CreatedAt:  p.CreatedAt,
		Receptions: []ReceptionInfo{},
	}, nil
}

func (s *Service) List(ctx context.Context, from, to *time.Time, limit, offset int) ([]PVZInfo, error) {
	pvzList, err := s.pvzRepo.List(ctx, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	var result []PVZInfo
	for _, p := range pvzList {
		riList, err := s.receptionRepo.GetByPVZ(ctx, p.ID, from, to)
		if err != nil {
			return nil, err
		}
		var recvInfos []ReceptionInfo
		for _, r := range riList {
			piList, err := s.productRepo.GetByReception(ctx, r.ID)
			if err != nil {
				return nil, err
			}
			var prodInfos []ProductInfo
			for _, pr := range piList {
				prodInfos = append(prodInfos, ProductInfo{
					ID:      pr.ID,
					AddedAt: pr.AddedAt,
					Type:    pr.Type,
				})
			}
			recvInfos = append(recvInfos, ReceptionInfo{
				ID:        r.ID,
				StartedAt: r.StartedAt,
				Status:    r.Status,
				Products:  prodInfos,
			})
		}
		result = append(result, PVZInfo{
			ID:         p.ID,
			City:       p.City,
			CreatedAt:  p.CreatedAt,
			Receptions: recvInfos,
		})
	}
	return result, nil
}

func pvzCityEnglish(city string) string {
	switch city {
	case "Москва":
		return "Moscow"
	case "Санкт-Петербург":
		return "Saint Petersburg"
	case "Казань":
		return "Kazan"
	}
	return city
}

func normalizeCity(city string) (string, bool) {
	c := strings.TrimSpace(city)

	for _, allowed := range pvz.AllowedCities {
		if strings.EqualFold(c, allowed) {
			return allowed, true
		}
		if strings.EqualFold(c, pvzCityEnglish(allowed)) {
			return allowed, true
		}
	}
	return "", false
}
