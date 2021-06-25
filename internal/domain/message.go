package domain

import (
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrFailedToSaveMsg = errors.New("failed to save msg")

type Message struct {
	ID         int64                  `json:"id,omitempty"`
	Code       int64                  `json:"code,omitempty"`
	SenderID   int64                  `json:"sender_id,omitempty"`
	ReceiverID int64                  `json:"receiver_id,omitempty"`
	Time       *timestamppb.Timestamp `json:"time,omitempty"`
	Text       string                 `json:"text,omitempty"`
	InboxHash  int64                  `json:"inbox_hash,omitempty"`
}
