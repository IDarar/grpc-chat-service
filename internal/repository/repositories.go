package repository

import (
	"bytes"
	"database/sql"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	imagestore "github.com/IDarar/grpc-chat-service/internal/repository/image_store"
	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
)

type Messages interface {
	Save(msg *p.Message) error
}

type Repositories struct {
	Messages Messages
	Images   Images
}

type Images interface {
	Save(ext string, imageData bytes.Buffer) (string, error)
}

func NewRepositories(db *sql.DB, imageFolder string) *Repositories {
	return &Repositories{
		Messages: mysql.NewMessagesRepo(db),
		Images:   imagestore.NewDiskImageStore(imageFolder),
	}
}
