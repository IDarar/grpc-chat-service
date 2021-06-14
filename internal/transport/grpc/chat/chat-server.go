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

	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	mx      sync.RWMutex //due to connections are in map, there is needed an safe concurent access to map
	Service services.Services
	Cfg     *config.Config
}

//map contains clients's connections
//TODO think another way of getting connections to send msgs
var conns map[int64]p.ChatService_ConnectServer

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

	conns = make(map[int64]p.ChatService_ConnectServer)

	var messageServer ChatServer

	p.RegisterChatServiceServer(server, &messageServer)

	listen, err := net.Listen("tcp", ":"+s.Cfg.GRPC.Port)
	if err != nil {
		fmt.Println(err)
		return err
	}

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

	s.mx.Lock()
	conns[int64(sID)] = out
	s.mx.Unlock()

	errChan := make(chan error)

	//run goroutin handling msgs
	go s.receiveMsgs(out, errChan, int64(sID))

	//block until there is an error
	return <-errChan
}

func (s *ChatServer) GetMessages(ctx context.Context, req *p.RequestChatHistory) (*p.ChatHistory, error) {
	return nil, nil
}
func (s *ChatServer) GetInboxes(ctx context.Context, req *p.RequestInboxes) (*p.Chats, error)

func (s *ChatServer) receiveMsgs(stream p.ChatService_ConnectServer, errChan chan error, uID int64) {
	for {
		select {
		case <-errChan:
			return
		default:
			res, err := stream.Recv()
			if err != nil {
				logger.Error(err)
				errChan <- err
				return
			}
			go s.Service.Messages.Save(res)

			logger.Info("received msg to ", &res.ReceiverID)

			s.mx.RLock()
			destination, ok := conns[res.ReceiverID]
			s.mx.RUnlock()
			if ok {
				res.SenderID = uID

				//I think it will better if client side will handle creating inboxes. Otherwise it is needed to check if inbox exists on each message
				err := destination.Send(res)

				if status.Code(err) == codes.Unavailable {
					logger.Error(err)

					s.mx.Lock()
					delete(conns, res.ReceiverID)
					s.mx.Unlock()
				}
			} else {
				logger.Info("user is not connected")
			}

		}
	}
}
