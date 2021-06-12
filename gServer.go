package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
}

var errEmptyID = errors.New("no ID in context")

var port = ":7777"

var conns map[string]p.ChatService_ConnectServer

const userIDctx = "user_id"

func main() {
	defer recover()

	server := grpc.NewServer()

	conns = make(map[string]p.ChatService_ConnectServer)

	var messageServer ChatServer

	p.RegisterChatServiceServer(server, messageServer)

	listen, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Serving requests...")
	server.Serve(listen)
}

func (ChatServer) Connect(out p.ChatService_ConnectServer) error {
	defer recover()

	res, err := out.Recv()
	if err != nil {
		logger.Error(err)
		return err
	}

	md, ok := metadata.FromIncomingContext(out.Context())
	logger.Info(md)
	if !ok {
		logger.Error(errEmptyID)
		return errEmptyID
	}

	senderID := md.Get(userIDctx)[0]

	conns[senderID] = out

	logger.Info(conns)

	logger.Info("greet message is ", res)

	errChan := make(chan error)

	go receiveMsgs(out, errChan, senderID)

	return <-errChan
}

func (ChatServer) GetMessages(ctx context.Context, req *p.RequestChatHistory) (*p.ChatHistory, error) {
	return nil, nil
}

func receiveMsgs(stream p.ChatService_ConnectServer, errChan chan error, uID string) {
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
			destination, ok := conns[res.ReceiverID]
			if ok {
				res.SenderID = uID
				err := destination.Send(res)
				if status.Code(err) == codes.Unavailable {
					logger.Error(err)
					delete(conns, res.ReceiverID)
				}
			} else {
				logger.Info("user is not connected")
			}

		}
	}
}
