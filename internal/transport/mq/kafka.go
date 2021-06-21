package mq

import (
	"time"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const chatTopic = "chat-messages"

func NewKafkaDialer(cfg *config.Config) *kafka.Dialer {

	mechanism := plain.Mechanism{
		Username: cfg.Kafka.UsernameSASL,
		Password: cfg.Kafka.PaswordSASL,
	}

	dialer := &kafka.Dialer{
		Timeout:       5 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	conn, err := dialer.Dial("tcp", cfg.Kafka.Host)
	if err != nil {
		panic(err.Error())
	}

	topicConfigs := kafka.TopicConfig{
		Topic:             chatTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = conn.CreateTopics(topicConfigs)
	if err != nil {
		panic(err.Error())
	}

	return dialer
}
