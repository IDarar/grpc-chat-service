package domain

//Maybe it doesn't make sense there, but intresting to try
/*
const MessagePoolSize = 800

var pool chan *p.Message

func init() {
	pool = make(chan *p.Message, MessagePoolSize)
}

func Alloc() *p.Message {
	select {
	case m := <-pool:
		return m
	default:
		m := &p.Message{}

		logger.Info("allocating msg")

		return m
	}
}

func Release(m *p.Message) {
	select {
	case pool <- m:
	default:
	}
}
*/
