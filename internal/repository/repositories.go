package repository

import (
	"bytes"
	"database/sql"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	imagestore "github.com/IDarar/grpc-chat-service/internal/repository/image_store"
	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
)

//go:generate mockgen -source=repositories.go -destination=mock_repository/repo_mocks.go

type Messages interface {
	Save(msg *p.Message) error
}

type Images interface {
	Save(ext string, imageData *bytes.Buffer) (string, error)
}

type Repositories struct {
	Messages Messages
	Images   Images
}

func NewRepositories(db *sql.DB, imageFolder string) *Repositories {
	return &Repositories{
		Messages: mysql.NewMessagesRepo(db),
		Images:   imagestore.NewDiskImageStore(imageFolder),
	}
}
