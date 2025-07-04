package server

import (
	"context"
	"gRPC-gateway/config"
	protogen "gRPC-gateway/internal/services/genproto/logs"
	"gRPC-gateway/internal/services/log_service"
	"log"
	"net"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)

type Kfk struct {
	Producer        sarama.SyncProducer
	ProtoSerializer *config.ProtobufSerializer
}

func StartNewgRPCServer(ctx context.Context, cfg *config.AppConfig, kfk *Kfk) error {
	lis, err := net.Listen("tcp", cfg.ServerPort)

	if err != nil {
		return err
	}
	s := grpc.NewServer()

	log.Println("Server started on port", cfg.ServerPort)

	logService := log_service.NewLogServiceServer(kfk.Producer, kfk.ProtoSerializer)
	protogen.RegisterLogServiceServer(s, logService)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()
	return s.Serve(lis)
}
