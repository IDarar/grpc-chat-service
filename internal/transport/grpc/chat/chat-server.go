package chat

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"github.com/IDarar/grpc-chat-service/internal/domain"
	"github.com/IDarar/grpc-chat-service/pkg/tlscredentials"

	"github.com/IDarar/hub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ChatServerRun(s *ChatServer) error {
	defer recover()

	//enable tls
	tlsCredentials, err := tlscredentials.LoadTLSCredentialsServer(&s.Cfg)
	if err != nil {
		return fmt.Errorf("cannot load TLS credentials: %w", err)
	}

	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.Creds(tlsCredentials))

	s.mx = sync.RWMutex{}

	server := grpc.NewServer(opts...)

	userConns = make(map[int64]*UserConnections)

	p.RegisterChatServiceServer(server, s)

	listen, err := net.Listen("tcp", ":"+s.Cfg.GRPC.Port)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//run checking connections before start server
	go s.ping()

	//run reading messages from mq
	go s.receiveMsgsFromMQ(userConns)

	fmt.Println("Serving requests...")
	return server.Serve(listen)
}

//On first request from client we store connection, on further requests handle messages
func (s *ChatServer) Connect(out p.ChatService_ConnectServer) error {
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

	//initalise connection object
	errChan := make(chan error, 1)

	s.addConnection(sID, &out, errChan)

	//handle messages from connection
	for {
		select {
		//exit if there is an errror
		case err = <-errChan:
			if err == domain.ErrFailedToSaveMsg {
				logger.Error("err saving ", err)
				out.Send(&p.Message{Code: Failed})

				continue
			}
			logger.Info("err on chan, exit goroutine ", err)

			//remove connection
			return err
		//otherwise handle messages from grpc connections
		default:
			res, err := out.Recv()
			if err != nil {
				logger.Error(err)
				errChan <- err
				return err
			}

			//if message contains images
			if len(res.Images) != 0 {
				logger.Info("num of images: ", len(res.Images))
				//save images and get kist of ids
				s.Service.Messages.Save(res, errChan)
				if err != nil {
					//if error during saving, send code Failed and restart loop
					logError(err)
					out.Send(&p.Message{Code: Failed})

					continue
				}
				//get the connection from map
				s.mx.RLock()
				destination, ok := userConns[res.ReceiverID]
				s.mx.RUnlock()

				//if user is not connected to this server
				//try  to send msg to mq, and maybe other instances handle needed connection
				if !ok {
					{
						logger.Info("user is not connected to this server, send msg to MQ")
						err = s.MQ.ChatMQ.WriteMessages(res)
						if err != nil {
							logger.Error("err writing to kafka ", err)

							continue
						}
					}
				} else {
					logger.Info("sending to ", res.ReceiverID)
					s.sendMsg(res, destination)
					continue
				}
				continue
			}

			logger.Info(res.Text)

			res.SenderID = int64(sID)

			logger.Info("received msg from ", res.SenderID)

			logger.Info("received msg to ", res.ReceiverID, " on gRPC")

			//save msg asyncroniously
			go s.Service.Messages.Save(res, errChan)

			//get the connection from map
			s.mx.RLock()
			destination, ok := userConns[res.ReceiverID]
			s.mx.RUnlock()

			//if user is not connected to this server
			//try  to send msg to mq, and maybe other instances handle needed connection
			if !ok {
				{
					logger.Info("user is not connected to this server, send msg to MQ")
					err = s.MQ.ChatMQ.WriteMessages(res)
					if err != nil {
						logger.Error("err writing to kafka ", err)

						continue
					}
				}
			} else {
				logger.Info("sending to ", res.ReceiverID)
				s.sendMsg(res, destination)
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
func (s *ChatServer) receiveMsgsFromMQ(conns map[int64]*UserConnections) {
	for {

		msg, err := s.MQ.ChatMQ.ReadMessages()
		if err != nil {
			logger.Error(err)
			continue
		}
		if msg == nil {
			logger.Info("nil message on mq")
			continue
		}
		logger.Info("received msg to ", msg.ReceiverID, " on MQ")

		//get connection with user
		s.mx.RLock()
		uConns, ok := conns[msg.ReceiverID]
		s.mx.RUnlock()

		if !ok {
			continue
		}

		//send msg
		s.sendMsgConnection(uConns, msg)
	}
}

func (s *ChatServer) addConnection(sID int, out *p.ChatService_ConnectServer, errChan chan error) {
	//check if server has conenction with user
	s.mx.RLock()
	conns, ok := userConns[int64(sID)]
	s.mx.RUnlock()

	//initalise connection object

	//if dont have, create new struct containg all users connections
	if !ok {
		newConn := &UserConnections{ID: sID}
		newConn.conns = append(newConn.conns, &Connection{conn: *out, errChan: errChan})
		userConns[int64(sID)] = newConn

		logger.Info("user with id ", sID, " joins to chat")

		// if user already has other connections with this erver, append slice of connections
	} else {
		conns.conns = append(conns.conns, &Connection{conn: *out, errChan: errChan})

		logger.Info("user with id ", sID, " have other connections to server")
	}
}

func (s *ChatServer) sendMsg(msg *p.Message, uConns *UserConnections) {
	if len(uConns.conns) == 1 {
		logger.Info("send to only connection ...")
		err := uConns.conns[0].conn.Send(msg)
		if err != nil {
			logger.Info("only user's connection does not respond, remove from map")

			//sendin fails, delete conenctions
			s.mx.Lock()
			delete(userConns, int64(uConns.ID))
			s.mx.Unlock()
		}

		//if len of user's connections is bigger than 1
	} else if len(uConns.conns) > 1 {
		//number of user's connections that will be decremented if coonection does not respond
		//if it is 0 than user dont have active conns, so remove it from map
		activeConns := len(uConns.conns)
		logger.Info("send to all connections ...")

		for i := len(uConns.conns) - 1; i >= 0; i-- {
			err := uConns.conns[i].conn.Send(msg)
			if err != nil {

				//when connection from user brokes the Connect function exits
				//so there is no goroutine that can read from that channel
				//and writing to this channel blocks execution of function

				//v.errChan <- err

				// decrement num of active conns
				activeConns -= 1

				logger.Info("deleting connection from slice of user ", uConns.ID)

				//remove connection from slice
				uConns.conns[i] = uConns.conns[len(uConns.conns)-1]
				uConns.conns[len(uConns.conns)-1] = nil
				uConns.conns = uConns.conns[:len(uConns.conns)-1]
			}
		}

		//no active conenctions, remove struct
		if activeConns == 0 {
			logger.Info("deleting whole connection")
			delete(userConns, int64(uConns.ID))
		}
		//if 0
	} else {
		logger.Info("deleting whole connection")
		delete(userConns, int64(uConns.ID))
	}
}

func logError(err error) error {
	if err != nil {
		logger.Error(err)
	}
	return err
}
