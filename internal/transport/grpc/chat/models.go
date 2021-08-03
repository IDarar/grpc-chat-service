package chat

import (
	"reflect"
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/mq"
	"github.com/IDarar/hub/pkg/logger"
)

type ChatServer struct {
	mx      sync.RWMutex //due to connections are in map, there is needed an safe concurent access to map
	Service services.Services
	MQ      mq.MQ
	Cfg     config.Config
}

//All user's connections will be definetly only on one instance of app
//that will be garanued by load balaner
//So it is not needed to send many abundant messages to MQ in order to ensure
//that message will be delivered to user
type UserConnections struct {
	ID    int
	conns []*Connection
}

type Connection struct {
	conn    p.ChatService_ConnectServer
	errChan chan error
}

//map contains array of clients's connections
//TODO think another way of getting connections to send msgs
//MQ there is only solution I see, for, if app will be scaled for more than 1 instance, conns will be on different servers, so it will cause problems.
//array of connections is needed because thtere is only id with which needed user can be found, but he can have multiple connections
var userConns map[int64]*UserConnections

const userIDctx = "user_id"

func NewServer(cfg *config.Config, services services.Services, mq mq.MQ) *ChatServer {
	return &ChatServer{
		Service: services,
		Cfg:     *cfg,
		MQ:      mq,
		mx:      sync.RWMutex{},
	}
}

//send msg, and remove connection if there is an error
func (s *ChatServer) sendMsgConnection(uConns *UserConnections, msg *p.Message) {
	if len(uConns.conns) == 1 {
		err := uConns.conns[0].conn.Send(msg)
		if err != nil {
			uConns.conns[0].errChan <- err
			logger.Info("only user connection does not respond, remove from map")

			//sendin fails, delete conenctions
			s.mx.Lock()
			delete(userConns, msg.ReceiverID)
			s.mx.Unlock()
		}
	}

	activeConns := len(uConns.conns)

	for j, v := range uConns.conns {
		err := v.conn.Send(msg)
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

	//no active conenctions, remove struct
	if activeConns == 0 {
		s.mx.Lock()
		delete(userConns, msg.ReceiverID)
		s.mx.Unlock()
	}
}

//get file descriptor of grpc stream
//dont hane sense untill I know how to preserve connections
//with returning from function
func streamFD(srvStream p.ChatService_ConnectServer) int {
	ServerStream := reflect.Indirect(reflect.ValueOf(srvStream)).FieldByName("ServerStream").Elem().Elem()
	stream := reflect.Indirect(ServerStream).FieldByName("s").Elem()
	streamTransport := reflect.Indirect(stream).FieldByName("st").Elem()
	conn := reflect.Indirect(streamTransport).FieldByName("conn").Elem()
	fdVall := reflect.Indirect(conn).FieldByName("fd")
	pfdVall := reflect.Indirect(fdVall).FieldByName("pfd")

	return int(pfdVall.FieldByName("Sysfd").Int())
}
