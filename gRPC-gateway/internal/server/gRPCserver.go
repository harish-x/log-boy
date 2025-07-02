package server

import (
	"context"
	protogen "gRPC-gateway/internal/services/genproto/logs"
	"gRPC-gateway/internal/services/log_service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func StartNewgRPCServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		return err
	}
	s := grpc.NewServer()

	log.Println("Server started")
	logService := log_service.NewLogServiceServer()
	protogen.RegisterLogServiceServer(s, logService)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()
	return s.Serve(lis)
}
