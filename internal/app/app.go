package app

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/repository"
	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/grpc/chat"

	"github.com/IDarar/hub/pkg/logger"
)

type network struct {
}

func (n *network) Network() string {
	return "tcp"
}

func (n *network) String() string {
	return "localhost:49157"
}

func Run(configPath string) {
	envInit()

	cfg, err := config.Init(configPath)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(cfg)

	db, err := mysql.NewMySQLDB(cfg)
	if err != nil {
		logger.Error(err)
		return
	}

	defer db.Close()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error(err)
		return
	}

	res, err := db.Exec("INSERT INTO messages (inbox_hash, created_at, sender_id, text) VALUES (?, ?, ?, ?)", 450456, time.Now(), 45, "msg.Text")

	if err != nil {
		logger.Error(err)
		tx.Rollback()
		return
	}
	logger.Info(res)

	err = tx.Commit()
	if err != nil {
		logger.Error(err)
		return
	}

	repos := repository.NewRepositories(db)

	services := services.NewServices(services.Deps{Repos: repos})

	chatServer := chat.NewServer(cfg, *services)

	log.Fatal(chatServer.ChatServerRun())
}

func envInit() {
	e := flag.Bool("env", false, "run app local?")

	flag.Parse()
	logger.Info("deploy env is ", *e)

	if *e {
		os.Setenv("MYSQL_PORT", "3306")
		os.Setenv("MYSQL_DATABASE", "chat")
		os.Setenv("MYSQL_PASSWORDGO", "secret")

		os.Setenv("KAFKA_USERSASL", "admin")
		os.Setenv("KAFKA_PASSWORDSASL", "admin-secret")
		os.Setenv("KAFKA_HOST", "localhost:9092")
		os.Setenv("KAFKA_NUMPARTITIONS", "1")
		os.Setenv("KAFKA_REPLICATIONFACTOR", "1")

		//TODO env for test db
		//os.Setenv("DATABASE_URL", "user=postgres dbname=hub_tests password=123 sslmode=disabled")
	}
}
