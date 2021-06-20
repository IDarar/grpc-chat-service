package app

import (
	"flag"
	"log"
	"os"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/grpc-chat-service/internal/repository"
	"github.com/IDarar/grpc-chat-service/internal/repository/mysql"
	"github.com/IDarar/grpc-chat-service/internal/services"
	"github.com/IDarar/grpc-chat-service/internal/transport/grpc/chat"
	"github.com/IDarar/grpc-chat-service/internal/transport/mq"

	"github.com/IDarar/hub/pkg/logger"
)

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

	repos := repository.NewRepositories(db)

	services := services.NewServices(services.Deps{Repos: repos})

	dialer := mq.NewKafkaDialer(cfg)
	cmq := mq.NewMQ(dialer, cfg)

	chatServer := chat.NewServer(cfg, *services, cmq)

	//_, err = chatServer.MQ.ChatMQ.ReadMessages(1)

	if err != nil {
		logger.Error(err)
		return
	}

	log.Fatal(chat.ChatServerRun(chatServer))
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
