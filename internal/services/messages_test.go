package services

import (
	"bytes"
	"testing"

	imagestore "github.com/IDarar/grpc-chat-service/internal/repository/image_store"
	"github.com/IDarar/grpc-chat-service/internal/repository/mock_repository"
	"github.com/IDarar/hub/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/IDarar/grpc-chat-service/chat_service"
)

func TestSaveImage(t *testing.T) {
	testcases := []struct {
		name   string
		images []*chat_service.Image
	}{
		{
			name: "successful saving one image",
			images: []*chat_service.Image{
				{
					ImageType: "png",
					ChankData: []byte("some bytes"),
				},
			},
		},
		{
			name: "successful saving two images",
			images: []*chat_service.Image{
				{
					ImageType: "png",
					ChankData: []byte("some bytes"),
				},
				{
					ImageType: "jpg",
					ChankData: []byte("some other bytes"),
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			rImages := mock_repository.NewMockImages(c)

			calls := []*gomock.Call{}
			for _, v := range tc.images {
				buf := &bytes.Buffer{}
				buf.Write(v.ChankData)

				calls = append(calls, rImages.EXPECT().Save(v.ImageType, buf).Return(imagestore.RandomString()+"."+v.ImageType, nil))
			}

			gomock.InOrder(calls...)

			s := MessagesService{imageStore: rImages}

			s.SaveImage(tc.images)

			for _, v := range tc.images {
				logger.Info(v.ImageID)
				assert.NotZero(t, v.ImageID)
				assert.Equal(t, 19+len(v.ImageType), len(v.ImageID))
			}
		})
	}

}
