package mq

import (
	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/segmentio/kafka-go"
)

type ChatMQ interface {
	ReadMessages() (*p.Message, error)
	WriteMessages(msg *p.Message) error
}

type MQ struct {
	ChatMQ ChatMQ
}

func NewMQ(dialer *kafka.Dialer, cfg *config.Config) MQ {
	return MQ{ChatMQ: NewChatMQ(dialer, *cfg)}
}
