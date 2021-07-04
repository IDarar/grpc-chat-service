package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

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
			if len(res.Images) != 0 {
				logger.Info("You received a message from ", res.SenderID, " text is ", res.Text, " code is ", res.Code, "images ids: ", res.Images)

				continue
			}

			logger.Info("You received a message from ", res.SenderID, " text is ", res.Text, " code is ", res.Code)

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

			fmt.Println("Enter ID of who you want to send a message ...")
			scanner.Scan()

			receiver := scanner.Text()

			rID, err := strconv.Atoi(receiver)
			if err != nil {
				logger.Error(err)
				return
			}

			fmt.Println("Enter text of your message ...")
			scanner.Scan()

			text := scanner.Text()

			fmt.Println("Number of images to message ...")
			scanner.Scan()

			num := scanner.Text()

			numInt, err := strconv.Atoi(num)
			if err != nil {
				logger.Error(err)
				return
			}
			msg := p.Message{ReceiverID: int64(rID), Text: text, Time: timestamppb.Now()}
			//send images
			if numInt != 0 {

				for i := 0; i < numInt; i++ {
					/*file, err := os.Open("./1.png")
					if err != nil {
						log.Fatal("cannot open image file: ", err)
					}
					defer file.Close()*/

					//buffer := []byte{}

					n, err := os.ReadFile("./1.png")

					if err != nil {
						log.Fatal("cannot read image file: ", err)
					}

					logger.Info("num of bytes: ", len(n))

					msg.Images = append(msg.Images, &p.Image{ImageType: "png", ChankData: n})
				}

				err = stream.Send(&msg)
				if err != nil {
					logger.Error(err)
					errChan <- err
					return
				}
				logger.Info("sended")

				continue
			}

			err = stream.Send(&msg)
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

	fmt.Println("Please input your ID to enter the chat")

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
