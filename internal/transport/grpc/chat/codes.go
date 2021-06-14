package chat

import "errors"

var (
	errEmptyID = errors.New("no ID in context")
)

const (
	Greet    = 0
	Message  = iota
	Close    = iota
	Sended   = iota
	Failed   = iota
	NewInbox = iota
)
