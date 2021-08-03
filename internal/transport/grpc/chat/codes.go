package chat

import "errors"

var (
	errEmptyID = errors.New("no ID in context")
)

const (
	Greet = iota
	Message
	Close
	Sended
	Failed
	NewInbox
	Ping
)
