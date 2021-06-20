package services

import (
	"github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/repository"
)

type Messages interface {
	Save(*chat_service.Message, chan error)
}

type Services struct {
	Messages Messages
}

type Deps struct {
	Repos *repository.Repositories
}

func NewServices(deps Deps) *Services {
	messages := NewMessagesService(deps.Repos.Messages)
	return &Services{Messages: messages}
}
