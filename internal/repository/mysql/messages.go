package mysql

import (
	"context"
	"database/sql"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/domain"
	"github.com/IDarar/hub/pkg/logger"
)

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{
		db: db,
	}
}
func (r *MessagesRepo) Save(msg *p.Message) {
	defer domain.Release(msg)

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error(err)
		return
	}

	goMsgTime := msg.Time.AsTime()

	res, err := r.db.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text) VALUES (?, ?, ?, ?)", msg.InboxHash, goMsgTime, msg.SenderID, msg.Text)
	if err != nil {
		logger.Error(err)
		tx.Rollback()
		return
	}
	logger.Info(res)

	err = tx.Commit()
	if err != nil {
		logger.Error(err)
		return
	}
}
