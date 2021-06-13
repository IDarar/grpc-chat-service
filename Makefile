run:
	go run ./cmd/chat-server/main.go --env
startcompose:
	sudo docker-compose up -d	
entermysql:
	sudo docker exec -it mysql mysql -u root -p
upmigrate:
	migrate -path db-migrate -database "mysql://root:secret@tcp(127.0.0.1:3306)/chat" -verbose up
downmigrate:
	migrate -path db-migrate -database "mysql://root:secret@tcp(127.0.0.1:3306)/chat" -verbose down	
proto:
	protoc -I . api.proto --go_out=plugins=grpc:.