package chat

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/domain"

	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

	//run reading messages from mq
	go s.receiveMsgsFromMQ(conns)

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

	//get user ID from metadata
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

	//handle messages from connection
	for {
		select {
		//exit if there is an errror
		case err = <-chatConn.errChan:
			if err == domain.ErrFailedToSaveMsg {
				logger.Error("err saving ", err)
				chatConn.conn.Send(&p.Message{Code: Failed})

				continue
			}
			return err
		//otherwise handle messages from grpc connections
		default:
			res, err := chatConn.conn.Recv()
			if err != nil {
				logger.Error(err)
				chatConn.errChan <- err
				return err
			}

			logger.Info(res)

			res.SenderID = int64(sID)

			logger.Info("received msg from ", res.SenderID)

			logger.Info("received msg to ", res.ReceiverID, " on gRPC")

			//save msg asyncroniously
			go s.Service.Messages.Save(res, chatConn.errChan)

			//get the connection from map
			s.mx.RLock()
			destination, ok := conns[res.ReceiverID]
			s.mx.RUnlock()

			//if exists send msg directly
			if ok {
				res.SenderID = int64(sID)
				err := destination.conn.Send(res)
				if status.Code(err) == codes.Unavailable {
					logger.Error(err)
					delete(conns, res.ReceiverID)
				}
				//otherwise user is not connected to this server
				//so last try is to send msg to mq, and maybe other instances handle needed connection
				//or user is not cconnected
			} else {
				logger.Info("user is not connected to this server, send msg to MQ")
				err = s.MQ.ChatMQ.WriteMessages(res)
				if err != nil {
					logger.Error("err writing to kafka ", err)
					chatConn.errChan <- err
					return err
				}
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

//for it is much more efficient to see in map if connection exists than to create reader for every connection
//there will be on reader for all connections. Further there can be many readers for particular amount of connections.
func (s *ChatServer) receiveMsgsFromMQ(conns map[int64]*ChatConnection) {
	for {

		msg, err := s.MQ.ChatMQ.ReadMessages()
		if err != nil {
			logger.Error(err)
			logger.Error("error on mq, exit goroutine ...", err)
			continue
		}
		if msg == nil {
			logger.Info("nil message on mq")
			continue
		}
		logger.Info("received msg to ", msg.ReceiverID, " on MQ")

		s.mx.RLock()
		err = conns[msg.ReceiverID].conn.Send(msg)
		if err != nil {
			logger.Error(err)
			delete(conns, msg.ReceiverID)
		}
		s.mx.RUnlock()
	}
}
