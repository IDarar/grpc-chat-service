package mysql

import "database/sql"

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{
		db: db,
	}
}
func (r *MessagesRepo) Save() error {
	return nil
}
