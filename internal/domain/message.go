package domain

import (
	"time"
)

type Message struct {
	ID         int
	Code       int
	SenderID   string
	ReceiverID string
	Time       time.Time
	Text       string
}

type Inbox struct {
	ID   int
	Hash int //maybe it will be of some different type

}
