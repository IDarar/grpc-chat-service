package chat

import (
	"time"

	"github.com/IDarar/grpc-chat-service/chat_service"
)

//will ping connections and remove if they dont respond
func (s *ChatServer) ping() {
	//in some amount of time
	ticker := time.NewTicker(20 * time.Minute)

	for range ticker.C {
		//range map of connections
		for _, uConns := range userConns {
			//if user have only one etablished connections
			s.sendMsg(&chat_service.Message{Code: Ping}, uConns)
		}
	}
}
