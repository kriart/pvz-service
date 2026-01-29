package main

import (
	"fmt"
	"log"
	"net"

	"pvz-service/internal/adapter/db/postgres"
	clockad "pvz-service/internal/adapter/time"
	"pvz-service/internal/config"
	"pvz-service/internal/transport/grpc/handler"
	"pvz-service/internal/transport/grpc/pb"
	"pvz-service/internal/usecase/pvz"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	db, err := postgres.NewDB(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		log.Fatal("Failed to connect DB: ", err)
	}

	clock := clockad.RealClock{}
	pvzService := pvz.NewService(db.PVZRepo(), db.ReceptionRepo(), db.ProductRepo(), clock)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterPVZServiceServer(grpcServer, handler.NewPVZServer(pvzService))

	log.Printf("gRPC server started on port %d", cfg.Server.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("gRPC server error:", err)
	}
}
