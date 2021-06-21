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
	//defer domain.Release(msg)

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error("err begin tx ", err)
		return domain.ErrFailedToSaveMsg
	}

	goMsgTime := msg.Time.AsTime()

	logger.Info("saving ...")

	_, err = r.db.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text) VALUES (?, ?, ?, ?)", msg.InboxHash, goMsgTime, msg.SenderID, msg.Text)
	if err != nil {
		logger.Error("err inserting ", err)
		tx.Rollback()
		return domain.ErrFailedToSaveMsg
	}

	var lastMsg string

	if len(msg.Text) > 60 {
		lastMsg = msg.Text[:60]
	} else {
		lastMsg = msg.Text
	}

	res, err := r.db.Exec("UPDATE inboxes SET last_msg = ?, seen = ?, unseen_number = unseen_number + 1 WHERE inbox_hash = ? AND user_id = ?", lastMsg, 0, msg.InboxHash, msg.ReceiverID)
	if err != nil {
		logger.Error("err updating ", err)
		tx.Rollback()
		return domain.ErrFailedToSaveMsg
	}

	n, err := res.RowsAffected()
	if err != nil {
		logger.Error("err commiting ", err)
		return domain.ErrFailedToSaveMsg
	}

	logger.Info("Db affected ... ", n, " rows")

	err = tx.Commit()
	if err != nil {
		logger.Error("err commiting ", err)
		return domain.ErrFailedToSaveMsg
	}

	return nil
}
