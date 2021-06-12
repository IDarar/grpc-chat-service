run:
	go run ./cmd/chat-server/main.go --env
startcompose:
	sudo docker-compose up -d	
proto:
	protoc -I . api.proto --go_out=plugins=grpc:.