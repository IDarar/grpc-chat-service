package services

import (
	"bytes"

	"github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/repository"
	"github.com/IDarar/hub/pkg/logger"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessagesService struct {
	repo       repository.Messages
	imageStore repository.Images
}

func (s *MessagesService) Save(msg *chat_service.Message, errCh chan error) {
	//calculate inbox hash
	senderID := msg.SenderID * 10
	msg.InboxHash = senderID*padding(msg.ReceiverID) + msg.ReceiverID
	msg.Time = timestamppb.Now()

	if len(msg.Images) != 0 {
		err := s.SaveImage(msg.Images)
		if err != nil {
			errCh <- err
			return
		}
	}

	err := s.repo.Save(msg)
	if err != nil {
		errCh <- err
	}
}

//saves images and sets ids to each
func (s *MessagesService) SaveImage(imgs []*chat_service.Image) error {
	for _, v := range imgs {
		imageData := &bytes.Buffer{}

		_, err := imageData.Write(v.ChankData)
		if err != nil {
			return logError(err)
		}

		id, err := s.imageStore.Save(v.ImageType, imageData)
		if err != nil {
			return logError(err)
		}

		//random generating of ids for images depends on current time
		//that can't update itself to give random value
		//so TODO change way of generating random ids for images
		//time.Sleep(100 * time.Millisecond)

		v.ChankData = nil
		v.ImageID = id
	}

	return nil
}

func NewMessagesService(repo repository.Messages, imageStore repository.Images) *MessagesService {
	return &MessagesService{
		repo:       repo,
		imageStore: imageStore,
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

func logError(err error) error {
	if err != nil {
		logger.Error(err)
	}
	return err
}
