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

	//begin transaction
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error("err begin tx ", err)
		return domain.ErrFailedToSaveMsg
	}

	//convert to golang time
	goMsgTime := msg.Time.AsTime()

	logger.Info("saving ...")

	//to get last inserted id
	var res sql.Result

	//if msg contains images set flag to row in db
	if len(msg.Images) != 0 {
		res, err = tx.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text, related_data) VALUES (?, ?, ?, ?, ?)", msg.InboxHash, goMsgTime, msg.SenderID, msg.Text, 1)
		if err != nil {
			logger.Error("err inserting ", err)
			tx.Rollback()
			return domain.ErrFailedToSaveMsg
		}
	} else {
		res, err = tx.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text) VALUES (?, ?, ?, ?)", msg.InboxHash, goMsgTime, msg.SenderID, msg.Text)
		if err != nil {
			logger.Error("err inserting ", err)
			tx.Rollback()
			return domain.ErrFailedToSaveMsg
		}
	}

	//get last message for inbox by first symbols of text of message
	var lastMsg string

	if len(msg.Text) > 60 {
		lastMsg = msg.Text[:60]
	} else {
		lastMsg = msg.Text
	}

	//update inbox
	_, err = tx.Exec("UPDATE inboxes SET last_msg = ?, seen = ?, unseen_number = unseen_number + 1 WHERE inbox_hash = ? AND user_id = ?", lastMsg, 0, msg.InboxHash, msg.ReceiverID)
	if err != nil {
		logger.Error("err updating ", err)
		tx.Rollback()
		return domain.ErrFailedToSaveMsg
	}

	//save images if they are in message
	if len(msg.Images) != 0 {
		lastMsgID, err := res.LastInsertId()
		if err != nil {
			logger.Error("err saving images ", err)
			return domain.ErrFailedToSaveMsg
		}

		for _, v := range msg.Images {
			_, err = tx.Exec("INSERT INTO images (id, message_id, inbox_hash, sended_at ,sender_id) VALUES (?, ?, ?, ?)", v.ImageID, lastMsgID, goMsgTime, msg.InboxHash, goMsgTime, msg.SenderID)
			if err != nil {
				logger.Error("err saving images ", err)
				return domain.ErrFailedToSaveMsg
			}
		}
	}

	//end transaction
	err = tx.Commit()
	if err != nil {
		logger.Error("err commiting ", err)
		return domain.ErrFailedToSaveMsg
	}

	return nil
}
