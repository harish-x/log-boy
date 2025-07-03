package log_service

import (
	protogen "gRPC-gateway/internal/services/genproto/logs"
	"io"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	gokfkprotobuf "github.com/djedjethai/gokfk-regent/serde/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LogServiceServer struct {
	protogen.UnimplementedLogServiceServer
	producer        *kafka.Producer
	protoSerializer *gokfkprotobuf.Serializer
}

func NewLogServiceServer(kafka *kafka.Producer, protoSerializer *gokfkprotobuf.Serializer) *LogServiceServer {
	return &LogServiceServer{
		protogen.UnimplementedLogServiceServer{},
		kafka,
		protoSerializer,
	}
}

func (s *LogServiceServer) ReceiveLogsStream(stream grpc.ClientStreamingServer[protogen.Log, protogen.Response]) error {
	log.Println("New client stream connected")
	go func() {
		for e := range s.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()

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

		log.Printf("Received log: Service=%s, Level=%s, Message=%s\n",
			logMessage.GetServiceName(), logMessage.GetLevel(), logMessage.GetMessage())

		topic := logMessage.GetServiceName()
		if topic == "" {
			log.Println("Received log with empty serviceName, skipping Kafka production.")
			continue
		}

		if logMessage.Timestamp == nil {
			logMessage.Timestamp = timestamppb.Now()
		}

		// The Serialize method should be identical
		kafkaValue, err := s.protoSerializer.Serialize(topic, logMessage)
		if err != nil {
			log.Printf("Failed to serialize protobuf message for topic %s: %v", topic, err)
			continue
		}

		err = s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          kafkaValue,
		}, nil)
		if err != nil {
			log.Printf("Failed to produce message to Kafka for topic %s: %v", topic, err)
		}
	}
}
