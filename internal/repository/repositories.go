package repository

import (
	"database/sql"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
)

type Messages interface {
	Save(msg *p.Message)
}

type Repositories struct {
	Messages Messages
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{Messages: mysql.NewMessagesRepo(db)}
}
