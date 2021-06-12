package services

import (
	"github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/repository"
)

type MessagesService struct {
	repo *repository.Messages
}

func (s *MessagesService) Save(msg *chat_service.Message) error {
	return nil
}

func NewMessagesService(repo *repository.Messages) *MessagesService {
	return &MessagesService{
		repo: repo,
	}

}
