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
func (r *MessagesRepo) Save(msg *p.Message) error {
	defer domain.Release(msg)

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error(err)
		return domain.ErrFailedToSaveMsg
	}

	goMsgTime := msg.Time.AsTime()

	res, err := r.db.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text) VALUES (?, ?, ?, ?)", msg.InboxHash, goMsgTime, msg.SenderID, msg.Text)
	if err != nil {
		logger.Error(err)
		tx.Rollback()
		return domain.ErrFailedToSaveMsg
	}
	logger.Info(res)

	res, err = r.db.Exec("UPDATE inboxes SET last_msg = ?, seen = ?, unseen_number = unseen_number + 1 WHERE inbox_hash = ? AND user_id = ?", msg.Text[:70], 0, msg.InboxHash, msg.ReceiverID)

	if err != nil {
		logger.Error(err)
		tx.Rollback()
		return domain.ErrFailedToSaveMsg
	}
	logger.Info(res)

	err = tx.Commit()
	if err != nil {
		logger.Error(err)
		return domain.ErrFailedToSaveMsg
	}

	return nil
}
