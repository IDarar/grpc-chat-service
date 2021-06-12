package main

import "github.com/IDarar/grpc-chat-service/internal/app"

const configPath = "configs/main"

func main() {
	app.Run(configPath)
}
