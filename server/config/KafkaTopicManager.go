package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

type KafkaTopicManager struct {
	admin  sarama.ClusterAdmin
	config *sarama.Config
}

func NewKafkaTopicManager(brokers []string) (*KafkaTopicManager, error) {
	config := sarama.NewConfig()
	config.Version = sarama.MaxVersion

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster admin: %w", err)
	}

	return &KafkaTopicManager{
		admin:  admin,
		config: config,
	}, nil
}

func (ktm *KafkaTopicManager) CreateProjectTopic(projectName string) error {
	topicName := fmt.Sprintf("logs-%s", projectName)

	// Check if a topic already exists
	exists, err := ktm.topicExists(topicName)
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if exists {
		log.Printf("Topic %s already exists", topicName)
		return nil
	}

	// Create a topic
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
		ConfigEntries: map[string]*string{
			"cleanup.policy": stringPtr("delete"),
			"retention.ms":   stringPtr("604800000"), // 7 days
			"segment.ms":     stringPtr("86400000"),  // 1 day
		},
	}

	err = ktm.admin.CreateTopic(topicName, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}
	// create topics for metrics
	topicName = fmt.Sprintf("metrics-%s", projectName)

	// Check if a topic already exists
	exists, err = ktm.topicExists(topicName)
	if err != nil {
		return fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if exists {
		log.Printf("Topic %s already exists", topicName)
		return nil
	}

	// Create a topic
	topicDetail = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
		ConfigEntries: map[string]*string{
			"cleanup.policy": stringPtr("delete"),
			"retention.ms":   stringPtr("604800000"), // 7 days
			"segment.ms":     stringPtr("86400000"),  // 1 day
		},
	}

	err = ktm.admin.CreateTopic(topicName, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}

	log.Printf("Successfully created topic: %s", topicName)
	return nil
}

func (ktm *KafkaTopicManager) topicExists(topicName string) (bool, error) {
	metadata, err := ktm.admin.DescribeTopics([]string{topicName})
	if err != nil {
		return false, err
	}

	for _, topicMeta := range metadata {
		if topicMeta.Name == topicName {
			if !errors.Is(topicMeta.Err, sarama.ErrNoError) {
				if errors.Is(topicMeta.Err, sarama.ErrUnknownTopicOrPartition) {
					return false, nil
				}
				return false, fmt.Errorf("topic metadata error: %v", topicMeta.Err)
			}
			return true, nil
		}
	}

	return false, nil
}

func (ktm *KafkaTopicManager) GetTopicsWithPrefix(prefix string) ([]string, error) {
	metadata, err := ktm.admin.DescribeTopics(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topics: %w", err)
	}

	var matchingTopics []string
	for _, topicMeta := range metadata {
		if !errors.Is(topicMeta.Err, sarama.ErrNoError) {
			log.Printf("Topic %s has error: %v", topicMeta.Name, topicMeta.Err)
			continue
		}

		if strings.HasPrefix(topicMeta.Name, prefix) {
			matchingTopics = append(matchingTopics, topicMeta.Name)
		}
	}

	return matchingTopics, nil
}

func (ktm *KafkaTopicManager) Close() error {
	return ktm.admin.Close()
}

func stringPtr(s string) *string {
	return &s
}
