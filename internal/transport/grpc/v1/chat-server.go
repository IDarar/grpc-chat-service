package v1

import (
	"context"
	"errors"
	"fmt"
	"net"
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
}

var errEmptyID = errors.New("no ID in context")

//map contains clients's connections
//TODO think another way of getting connections to send msgs
var conns map[string]p.ChatService_ConnectServer

const userIDctx = "user_id"

func (s *ChatServer) ChatServerRun(cfg *config.Config) {
	defer recover()

	server := grpc.NewServer()

	conns = make(map[string]p.ChatService_ConnectServer)

	var messageServer ChatServer

	p.RegisterChatServiceServer(server, &messageServer)

	listen, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Serving requests...")
	server.Serve(listen)
}

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

	s.mx.Lock()
	conns[senderID] = out
	s.mx.Unlock()

	errChan := make(chan error)

	//run goroutin handling msgs
	go s.receiveMsgs(out, errChan, senderID)

	//block until there is an error
	return <-errChan
}

func (s *ChatServer) GetMessages(ctx context.Context, req *p.RequestChatHistory) (*p.ChatHistory, error) {
	return nil, nil
}

func (s *ChatServer) receiveMsgs(stream p.ChatService_ConnectServer, errChan chan error, uID string) {
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

			logger.Info("received msg to ", &res.ReceiverID)

			s.mx.RLock()
			destination, ok := conns[res.ReceiverID]
			s.mx.RUnlock()
			if ok {
				res.SenderID = uID
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