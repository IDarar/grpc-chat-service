package app

import (
	"fmt"
	"net"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"google.golang.org/grpc"
	"gorm.io/gorm/logger"
)

func Run(configPath string) {
	cfg, err := config.Init(configPath)
	if err != nil {
		logger.Error(err)
		return
	}

	server := grpc.NewServer()

	//conns = make(map[string]p.ChatService_ConnectServer)

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
