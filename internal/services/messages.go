package services

import (
	"github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessagesService struct {
	repo repository.Messages
}

func (s *MessagesService) Save(msg *chat_service.Message, errCh chan error) {
	//calculate inbox hash
	senderID := msg.SenderID * 10
	msg.InboxHash = senderID*padding(msg.ReceiverID) + msg.ReceiverID
	msg.Time = timestamppb.Now()

	err := s.repo.Save(msg)
	if err != nil {
		errCh <- err
	}
}

func NewMessagesService(repo repository.Messages) *MessagesService {
	return &MessagesService{
		repo: repo,
	}

}

//concantenate two values to get inbox
func padding(n int64) int64 {
	var p int64 = 1
	for p < n {
		p *= 10
	}

	return p
}
