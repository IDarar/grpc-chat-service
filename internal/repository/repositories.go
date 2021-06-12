package repository

import (
	"database/sql"

	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
)

type Messages interface {
	Save() error
}

type Repositories struct {
	Messages Messages
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{Messages: mysql.NewMessagesRepo(db)}
}
