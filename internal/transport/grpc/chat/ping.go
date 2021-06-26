package chat

import (
	"time"

	"github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/hub/pkg/logger"
)

//will ping connections and remove if they dont respond
func (s *ChatServer) ping() {
	//in some amount of time
	ticker := time.NewTicker(20 * time.Minute)

	for range ticker.C {
		//range map of connections
		for i, uConns := range userConns {
			//if user have only one etablished connections
			if len(uConns.conns) == 1 {
				err := uConns.conns[0].conn.Send(&chat_service.Message{Code: Ping})
				if err != nil {
					uConns.conns[0].errChan <- err
					logger.Info("only user connection does not respond, remove from map")

					//sendin fails, delete conenctions
					s.mx.Lock()
					delete(userConns, i)
					s.mx.Unlock()

				}
				//if len of user's connections is bigger than 1
			} else if len(uConns.conns) > 1 {
				//number of user's connections that will be decremented if coonection does not respond
				//if it is 0 than user dont have active conns, so remove it from map
				activeConns := len(uConns.conns)

				for j, v := range uConns.conns {
					err := v.conn.Send(&chat_service.Message{Code: Ping})
					if err != nil {
						//send err to chan and decrement num of active conns
						v.errChan <- err
						activeConns -= 1

						//remove connection from slice
						uConns.conns[j] = uConns.conns[len(uConns.conns)-1]
						uConns.conns[len(uConns.conns)-1] = nil
						uConns.conns = uConns.conns[:len(uConns.conns)-1]
					}
				}
				//no active conenctions, remove strucet
				if activeConns == 0 {
					delete(userConns, i)
				}
			}
		}
	}
}
