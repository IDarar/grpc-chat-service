package mq

import (
	"net"
	"strconv"
	"time"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const chatTopic = "chat-messages"

func NewKafkaDialer(cfg *config.Config) *kafka.Dialer {

	mechanism := plain.Mechanism{
		Username: cfg.Kafka.UsernameSASL,
		Password: cfg.Kafka.PaswordSASL,
	}

	dialer := &kafka.Dialer{
		Timeout:       5 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	conn, err := dialer.Dial("tcp", cfg.Kafka.Host)
	if err != nil {
		panic(err.Error())
	}

	//Init kafka creating chat topic
	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn

	controllerConn, err = dialer.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             chatTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}

	return dialer
}
