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

	repos := repository.NewRepositories(db, "../../../images")

	services := services.NewServices(services.Deps{Repos: repos})

	dialer := mq.NewKafkaDialer(cfg)

	cmq := mq.NewMQ(dialer, cfg)

	chatServer := chat.NewServer(cfg, *services, cmq)

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
		os.Setenv("KAFKA_NUMPARTITIONS", "0")
		os.Setenv("KAFKA_REPLICATIONFACTOR", "0")
		os.Setenv("KAFKA_GROUPID", "chat_s1")
		//TODO env for test db
		//os.Setenv("DATABASE_URL", "user=postgres dbname=hub_tests password=123 sslmode=disabled")
	}
}

/*
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.Host},

		Topic:    "chat-messages",
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})

	w.Addr = kafka.TCP(cfg.Kafka.Host)

	err = w.WriteMessages(context.Background(), kafka.Message{Key: []byte("123"), Value: []byte("fasfajsfklaj")})
	if err != nil {
		logger.Error(err)
		return
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{cfg.Kafka.Host},
		Topic:     "chat-messages",
		Partition: cfg.Kafka.NumPartitions,
		MaxBytes:  10e6, // 10MB
		Dialer:    dialer,
	})

	m, err := r.ReadMessage(context.Background())
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info(m)





	//defer conn.DeleteTopics(mq.ChatTopic)
	/*w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.Host},
		//Async:    true,
		Topic:    "chat-messages",
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})

	w.Addr = kafka.TCP(cfg.Kafka.Host)

	for i := 0; i < 3; i++ {

		err = w.WriteMessages(context.Background(), kafka.Message{Key: []byte("123"), Value: []byte("fasfajsfklaj")})
		if err != nil {
			logger.Error(err)
			return
		}

		fmt.Println(i)
	}

	c2 := 0
	c1 := 0

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{cfg.Kafka.Host},
		Topic:     "chat-messages",
		Partition: cfg.Kafka.NumPartitions,
		MaxBytes:  10e6, // 10MB
		Dialer:    dialer,
		//	StartOffset: kafka.FirstOffset,
		GroupID: "25",
	})

	//r.SetOffset(-1)

	go func() {
		logger.Info("start 1 reader")
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				logger.Error(err)

			}
			c1++
			logger.Info(c1)

			r.CommitMessages(context.Background(), m)
			logger.Info("from 1 reader ", string(m.Value))
		}

	}()

	r2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{cfg.Kafka.Host},
		Topic:     "chat-messages",
		Partition: cfg.Kafka.NumPartitions,
		MaxBytes:  10e6, // 10MB
		Dialer:    dialer,
		//	StartOffset: kafka.FirstOffset,
		GroupID: "25",
	})

	//r2.SetOffset(-1)

	for {
		logger.Info("start 2 reader")
		m, err := r2.ReadMessage(context.Background())
		if err != nil {
			logger.Error(err)

		}
		c2++
		logger.Info(c2)
		//r.CommitMessages(context.Background(), m)

		logger.Info("from reader 2 ", string(m.Value))
	}

*/
