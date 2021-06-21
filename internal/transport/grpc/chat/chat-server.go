package chat

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/domain"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/mq"

	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func ChatServerRun(s *ChatServer) error {
	defer recover()

	s.mx = sync.RWMutex{}

	server := grpc.NewServer()

	conns = make(map[int64]*ChatConnection)

	p.RegisterChatServiceServer(server, s)

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

	logger.Info("user with id ", senderID, " joins to chat")

	//initalise connection object
	errChan := make(chan error)

	chatConn := &ChatConnection{ID: sID, conn: out, errChan: errChan}

	s.mx.Lock()
	conns[int64(sID)] = chatConn
	s.mx.Unlock()

	//handle messages from MQ
	go s.receiveMsgsFromMQ(chatConn)

	//handle messages from from connection
	for {
		select {
		case err = <-chatConn.errChan:
			if err == domain.ErrFailedToSaveMsg {
				logger.Error("err saving ", err)
				chatConn.conn.Send(&p.Message{Code: Failed})

				//TODO test how it will work with select
				break
			}
			return err
		default:
			res, err := chatConn.conn.Recv()
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return err
			}

			res.SenderID = int64(sID)
			logger.Info("received msg from ", res.SenderID)

			logger.Info("received msg to ", res.ReceiverID, " on gRPC")

			go s.Service.Messages.Save(res, chatConn.errChan)

			err = s.MQ.ChatMQ.WriteMessages(res)
			if err != nil {
				logger.Error("err writing to kafka ", err)
				chatConn.errChan <- err
				return err
			}

		}
	}
}

func (s *ChatServer) GetMessages(ctx context.Context, req *p.RequestChatHistory) (*p.ChatHistory, error) {
	return nil, nil
}
func (s *ChatServer) GetInboxes(ctx context.Context, req *p.RequestInboxes) (*p.Chats, error) {
	return nil, nil
}

func (s *ChatServer) CreateInbox(ctx context.Context, req *p.RequestCreateInbox) (*p.ChatHistory, error) {
	return nil, nil
}

func (s *ChatServer) receiveMsgsFromMQ(chatConn *ChatConnection) {
	for {
		select {
		case err := <-chatConn.errChan:
			if err == domain.ErrFailedToSaveMsg {
				logger.Error("err saving ", err)

				break
			}
			logger.Error("error on mq, exit goroutine ...")
			return
		default:
			msg, err := s.MQ.ChatMQ.ReadMessages(int(chatConn.ID))
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				logger.Error("error on mq, exit goroutine ...")
				return
			}
			if msg == nil {
				logger.Info("nil message on mq")
				break
			}
			logger.Info("received msg to ", msg.ReceiverID, " on MQ")

			err = chatConn.conn.Send(msg)
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return
			}
		}
	}
}
