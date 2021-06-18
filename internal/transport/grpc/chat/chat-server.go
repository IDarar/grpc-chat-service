package chat

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/mq"

	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ChatServer struct {
	mx      sync.RWMutex //due to connections are in map, there is needed an safe concurent access to map
	Service services.Services
	MQ      *mq.MQ
	Cfg     *config.Config
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

func NewServer(cfg *config.Config, services services.Services) *ChatServer {
	return &ChatServer{
		Service: services,
		Cfg:     cfg,
	}
}

func (s *ChatServer) ChatServerRun() error {
	defer recover()

	s.mx = sync.RWMutex{}

	server := grpc.NewServer()

	conns = make(map[int64]*ChatConnection)

	var messageServer ChatServer

	p.RegisterChatServiceServer(server, &messageServer)

	listen, err := net.Listen("tcp", ":"+s.Cfg.GRPC.Port)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//run checking connections before start server
	go ping()

	fmt.Println("Serving requests...")
	return server.Serve(listen)
}

//On first request from client we store connection, on further requests handle messages
func (s *ChatServer) Connect(out p.ChatService_ConnectServer) error {
	defer recover()

	_, err := out.Recv()
	if err != nil {
		logger.Error(err)
		return err
	}

	md, ok := metadata.FromIncomingContext(out.Context())
	if !ok {
		logger.Error(errEmptyID)
		return errEmptyID
	}

	senderID := md.Get(userIDctx)[0]

	sID, err := strconv.Atoi(senderID)
	if err != nil {
		logger.Error(err)
		return err
	}

	//initalise connection object
	errChan := make(chan error)

	chatConn := &ChatConnection{ID: sID, conn: out, errChan: errChan}

	s.mx.Lock()
	conns[int64(sID)] = chatConn
	s.mx.Unlock()

	//run goroutin handling msgs
	go s.receiveMsgsFromGrpc(chatConn)

	go s.receiveMsgsFromMQ(chatConn)

	//block until there is an error
	return <-errChan
}

func (s *ChatServer) GetMessages(ctx context.Context, req *p.RequestChatHistory) (*p.ChatHistory, error) {
	return nil, nil
}
func (s *ChatServer) GetInboxes(ctx context.Context, req *p.RequestInboxes) (*p.Chats, error) {
	return nil, nil
}

func (s *ChatServer) receiveMsgsFromGrpc(chatConn *ChatConnection) {
	for {
		select {
		case <-chatConn.errChan:
			return

		default:
			res, err := chatConn.conn.Recv()
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return
			}

			go s.Service.Messages.Save(res)

			logger.Info("received msg to ", &res.ReceiverID)

			err = s.MQ.ChatMQ.WriteMessages(res)
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return
			}

		}
	}
}

func (s *ChatServer) receiveMsgsFromMQ(chatConn *ChatConnection) {
	for {
		select {
		case <-chatConn.errChan:
			return

		default:
			msg, err := s.MQ.ChatMQ.ReadMessages(int(chatConn.ID))
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return
			}
			logger.Info("received msg to ", &msg.ReceiverID)

			err = chatConn.conn.Send(msg)
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return
			}
		}
	}
}
