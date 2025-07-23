package log_service

import (
	"gRPC-gateway/config"
	protogen "gRPC-gateway/internal/services/genproto/logs"
	"io"
	"log"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LogServiceServer struct {
	protogen.UnimplementedLogServiceServer
	producer        sarama.SyncProducer
	protoSerializer *config.ProtobufSerializer
}

func NewLogServiceServer(kafka sarama.SyncProducer, protoSerializer *config.ProtobufSerializer) *LogServiceServer {
	return &LogServiceServer{
		protogen.UnimplementedLogServiceServer{},
		kafka,
		protoSerializer,
	}
}

func (s *LogServiceServer) ReceiveLogsStream(stream grpc.ClientStreamingServer[protogen.Log, protogen.Response]) error {
	log.Println("New client stream connected")
	for {
		logMessage, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Client stream finished")
				return stream.SendAndClose(&protogen.Response{
					Ack: true,
				})
			}
			log.Printf("Failed to receive log: %v", err)
			return status.Errorf(codes.Unknown, "failed to receive log: %v", err)
		}

		topic := logMessage.GetServiceName()
		if topic == "" {
			log.Println("Received log with empty serviceName, skipping")
			continue
		}
		topic = "logs-" + topic
		log.Print("Producing log message to Kafka: ", logMessage)

		if logMessage.Timestamp == nil {
			logMessage.Timestamp = timestamppb.Now()
		}

		// Serialize protobuf message
		kafkaValue, err := s.protoSerializer.Serialize("Logs-value", logMessage)
		if err != nil {
			log.Printf("Failed to serialize protobuf message for topic %s: %v", topic, err)
			continue
		}

		// Create a Sarama producer message
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(kafkaValue),
		}

		// Send message synchronously
		partition, offset, err := s.producer.SendMessage(msg)
		if err != nil {
			log.Printf("Failed to produce message to Kafka for topic %s: %v", topic, err)
			continue
		}

		log.Printf("Delivered message to topic %s [%d] at offset %v\n",
			topic, partition, offset)
	}
}
