package chat

import (
	"context"
	"errors"
	"net"
	"reflect"
	"sync"
	"testing"

	p "github.com/IDarar/grpc-chat-service/chat_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestConnect(t *testing.T) {

}

type testSendMsgChatService_ConnectServer struct {
	grpc.ServerStream
	closed bool
}

//mock implementaion of server connections stream
func (t *testSendMsgChatService_ConnectServer) Send(*p.Message) error {
	if t.closed {
		return errors.New("connection closed")
	} else {
		return nil
	}
}
func (t *testSendMsgChatService_ConnectServer) Recv() (*p.Message, error) {
	return nil, nil

}

func TestSendMsg(t *testing.T) {
	s := ChatServer{mx: sync.RWMutex{}}

	testcases := []struct {
		name       string
		connection *UserConnections
		conns      map[int64]*UserConnections
		wantMap    map[int64]*UserConnections
	}{
		{
			name:       "one active connection",
			connection: &UserConnections{ID: 1, conns: []*Connection{{conn: &testSendMsgChatService_ConnectServer{closed: false}}}},
			conns:      map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{{conn: &testSendMsgChatService_ConnectServer{closed: false}}}}},
			wantMap:    map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{{conn: &testSendMsgChatService_ConnectServer{closed: false}}}}},
		},
		{
			name: "one active, one closed connection",
			connection: &UserConnections{ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}},
			conns: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}}},
			wantMap: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{{conn: &testSendMsgChatService_ConnectServer{closed: false}}}}},
		},
		{
			name: "one inactive",
			connection: &UserConnections{ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}},
			conns: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}}},
			wantMap: map[int64]*UserConnections{},
		},
		{
			name: "two inactive",
			connection: &UserConnections{ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}},
			conns: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}}},
			wantMap: map[int64]*UserConnections{},
		},
		{
			name: "two active, three closed",
			connection: &UserConnections{ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}},
			conns: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: true}},
			}}},
			wantMap: map[int64]*UserConnections{1: {ID: 1, conns: []*Connection{
				{conn: &testSendMsgChatService_ConnectServer{closed: false}},
				{conn: &testSendMsgChatService_ConnectServer{closed: false}}}}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			userConns = tc.conns

			s.sendMsg(&p.Message{Code: 0}, userConns[int64(tc.connection.ID)])

			eq := reflect.DeepEqual(userConns, tc.wantMap)
			if eq {
				return
			} else {
				t.Error("maps are not equal, logic of deletion conenctions does not work")
			}
			userConns = nil

		})
	}

}

func TestAddConnection(t *testing.T) {
	s := ChatServer{mx: sync.RWMutex{}}

	userConns = make(map[int64]*UserConnections)

	tests := []struct {
		name     string
		sID      int
		errChan  chan error
		out      p.ChatService_ConnectServer
		wantConn *UserConnections
	}{
		{
			name:     "first connection",
			sID:      1,
			out:      &testChatService_ConnectServer{},
			errChan:  make(chan error),
			wantConn: &UserConnections{ID: 1, conns: []*Connection{{}}},
		},
		{
			name:     "second connection",
			sID:      1,
			out:      &testChatService_ConnectServer{},
			errChan:  make(chan error),
			wantConn: &UserConnections{ID: 1, conns: []*Connection{{}, {}}},
		},
		{
			name:     "third connection",
			sID:      2,
			out:      &testChatService_ConnectServer{},
			errChan:  make(chan error),
			wantConn: &UserConnections{ID: 2, conns: []*Connection{{}}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			s.addConnection(tc.sID, &tc.out, tc.errChan)
			if len(userConns[int64(tc.sID)].conns) != len(tc.wantConn.conns) {
				t.Errorf("no connection in map")
			}
		})
	}
}

type testChatService_ConnectServer struct {
	grpc.ServerStream
}

func (*testChatService_ConnectServer) Send(*p.Message) error {
	return errors.New("")
}
func (*testChatService_ConnectServer) Recv() (*p.Message, error) {
	return nil, nil

}

func server(ctx context.Context) (p.ChatServiceClient, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	p.RegisterChatServiceServer(s, &ChatServer{mx: sync.RWMutex{}})
	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())

	closer := func() {
		listener.Close()
		s.Stop()
	}

	client := p.NewChatServiceClient(conn)

	return client, closer
}

func testNewCtxMD(id string) context.Context {
	ctx := context.Background()

	// add key-value pairs of metadata to context
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("user_id", id))

	return ctx
}
