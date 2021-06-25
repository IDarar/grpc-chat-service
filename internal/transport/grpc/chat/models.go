package chat

import (
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/mq"
)

type ChatServer struct {
	mx      sync.RWMutex //due to connections are in map, there is needed an safe concurent access to map
	Service services.Services
	MQ      mq.MQ
	Cfg     config.Config
}

type ChatConnection struct {
	ID      int
	conn    p.ChatService_ConnectServer
	errChan chan error
}

//map contains clients's connections
//TODO think another way of getting connections to send msgs
//MQ there is only solution I see, for, if app will be scaled for more than 1 instance, conns will be on different servers, so it will cause problems.
var conns map[int64]*ChatConnection

const userIDctx = "user_id"

func NewServer(cfg *config.Config, services services.Services, mq mq.MQ) *ChatServer {
	return &ChatServer{
		Service: services,
		Cfg:     *cfg,
		MQ:      mq,
	}
}
