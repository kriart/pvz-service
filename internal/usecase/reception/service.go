package reception

import (
	"context"

	"pvz-service/internal/domain/product"
	"pvz-service/internal/domain/reception"
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

func (s *Service) Open(ctx context.Context, pvzID string) (*reception.Reception, error) {
	p, err := s.pvzRepo.Get(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, reception.ErrNoOpenReception
	}

	openRec, err := s.receptionRepo.GetOpenByPVZ(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openRec != nil {
		return nil, reception.ErrReceptionAlreadyOpen
	}

	rec := &reception.Reception{
		ID:        uuid.New().String(),
		PVZID:     pvzID,
		StartedAt: s.clock.Now(),
		Status:    reception.StatusInProgress,
	}

	if err := s.receptionRepo.Create(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *Service) AddProduct(ctx context.Context, pvzID string, productType string) (*product.Product, error) {
	p, err := s.pvzRepo.Get(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, reception.ErrNoOpenReception
	}

	openRec, err := s.receptionRepo.GetOpenByPVZ(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openRec == nil {
		return nil, reception.ErrNoOpenReception
	}

	validType := false
	for _, t := range product.AllowedTypes {
		if t == productType {
			validType = true
			break
		}
	}
	if !validType {
		return nil, product.ErrInvalidType
	}

	prod := &product.Product{
		ID:          uuid.New().String(),
		ReceptionID: openRec.ID,
		AddedAt:     s.clock.Now(),
		Type:        productType,
	}

	if err := s.productRepo.Create(ctx, prod); err != nil {
		return nil, err
	}
	return prod, nil
}

func (s *Service) RemoveProduct(ctx context.Context, pvzID string) (*product.Product, error) {
	p, err := s.pvzRepo.Get(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, reception.ErrNoOpenReception
	}

	openRec, err := s.receptionRepo.GetOpenByPVZ(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openRec == nil {
		return nil, reception.ErrNoOpenReception
	}

	lastProd, err := s.productRepo.GetLastByReception(ctx, openRec.ID)
	if err != nil {
		return nil, err
	}
	if lastProd == nil {
		return nil, reception.ErrNoProducts
	}

	if err := s.productRepo.Delete(ctx, lastProd.ID); err != nil {
		return nil, err
	}
	return lastProd, nil
}

func (s *Service) Close(ctx context.Context, pvzID string) (*reception.Reception, error) {
	p, err := s.pvzRepo.Get(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, reception.ErrNoOpenReception
	}

	openRec, err := s.receptionRepo.GetOpenByPVZ(ctx, pvzID)
	if err != nil {
		return nil, err
	}
	if openRec == nil {
		return nil, reception.ErrNoOpenReception
	}

	if err := s.receptionRepo.Close(ctx, openRec.ID); err != nil {
		return nil, err
	}

	openRec.Status = reception.StatusClosed
	return openRec, nil
}
