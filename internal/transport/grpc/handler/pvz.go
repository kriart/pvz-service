package handler

import (
	"context"

	"pvz-service/internal/transport/grpc/pb"
	pvzuc "pvz-service/internal/usecase/pvz"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServer struct {
	pb.UnimplementedPVZServiceServer
	pvzService *pvzuc.Service
}

func NewPVZServer(pvzService *pvzuc.Service) *PVZServer {
	return &PVZServer{pvzService: pvzService}
}

func (s *PVZServer) GetPVZList(ctx context.Context, _ *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	list, err := s.pvzService.List(ctx, nil, nil, 100, 0)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetPVZListResponse{}
	for _, p := range list {
		resp.Pvzs = append(resp.Pvzs, &pb.PVZ{
			Id:               p.ID,
			RegistrationDate: timestamppb.New(p.CreatedAt),
			City:             p.City,
		})
	}
	return resp, nil
}
