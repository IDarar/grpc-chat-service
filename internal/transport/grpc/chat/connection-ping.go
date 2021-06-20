package chat

import (
	"time"

	"github.com/IDarar/grpc-chat-service/chat_service"
)

//will ping connections and remove if they dont respond
func ping() {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		for i, conn := range conns {
			err := conn.conn.Send(&chat_service.Message{Code: Ping})
			if err != nil {
				conn.errChan <- err
				delete(conns, i)
			}
		}
	}
}
