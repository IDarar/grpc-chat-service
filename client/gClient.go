package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	//
	p "github.com/IDarar/grpc-chat-service/chat_service"

	"github.com/IDarar/hub/pkg/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var port = ":7777"

const (
	Greet   = 0
	Message = iota
	Close   = iota
	Sended  = iota
	Failed  = iota
)

func main() {
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Dial:", err)
		return
	}
	defer conn.Close()
	client := p.NewChatServiceClient(conn)

	log.Fatal(Connect(client))
}

func Connect(client p.ChatServiceClient) error {
	ctx, err := getCtxWithName()
	if err != nil {
		logger.Error(err)

		return err
	}

	logger.Info("ctx is ", ctx)
	stream, err := client.Connect(ctx)
	if err != nil {
		logger.Error(err)

		return err
	}

	defer stream.CloseSend()

	errChan := make(chan error)

	greetReq := p.Message{Code: Greet, Time: timestamppb.Now()}

	err = stream.Send(&greetReq)
	if err != nil {
		logger.Error(err)

		return err
	}

	go receiveMsgs(stream, errChan)

	go sendMsgs(stream, errChan)

	//block until there is an error
	return <-errChan
}

func receiveMsgs(stream p.ChatService_ConnectClient, errChan chan error) {
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
			logger.Info("You received a message from ", res.SenderID, " text is ", res.Text)

		}
	}
}

func sendMsgs(stream p.ChatService_ConnectClient, errChan chan error) {
	for {
		select {
		case <-errChan:
			return
		default:

			scanner := bufio.NewScanner(os.Stdin)

			fmt.Println("Enter who you want to send a message ...")
			scanner.Scan()

			receiver := scanner.Text()

			fmt.Println("Enter text of your message ...")
			scanner.Scan()

			text := scanner.Text()
			msg := p.Message{ReceiverID: receiver, Text: text, Time: timestamppb.Now()}

			err := stream.Send(&msg)
			if err != nil {
				logger.Error(err)
				errChan <- err
				return
			}
			logger.Info("sended")
		}
	}
}

func getCtxWithName() (context.Context, error) {
	var name string

	fmt.Println("Please input your name to enter the chat")

	_, err := fmt.Fscan(os.Stdin, &name)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ctx := context.Background()

	// add key-value pairs of metadata to context
	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("user_id", name))

	return ctx, nil
}
