package mq

import (
	"context"
	"encoding/json"
	"strconv"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type ChatKafka struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

//Maybe later add custom partition r/w creation
func NewChatMQ(dialer *kafka.Dialer, cfg config.Config) *ChatKafka {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{cfg.Kafka.Host},
		Topic:    chatTopic,
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{cfg.Kafka.Host},
		Topic:     chatTopic,
		Partition: cfg.Kafka.NumPartitions,
		MaxBytes:  10e6, // 10MB
		Dialer:    dialer,
	})

	return &ChatKafka{writer: w, reader: r}
}

//gets messages from queue and checks if it is for needed user
func (k *ChatKafka) ReadMessages(uID int) (*p.Message, error) {
	m, err := k.reader.ReadMessage(context.Background())
	if err != nil {
		return nil, err
	}

	//key is ID of user
	key, err := strconv.Atoi(string(m.Key))
	if err != nil {
		return nil, err
	}
	if key != uID {
		return nil, nil
	}

	msg := domain.Alloc()
	err = json.Unmarshal(m.Value, msg)

	if err != nil {
		return nil, err
	}
	return msg, err
}

func (k *ChatKafka) WriteMessages(msg *p.Message) error {
	err := k.writer.WriteMessages(context.Background())
	domain.Release(msg)
	return err
}
