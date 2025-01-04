package incoming

import (
	"fmt"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/deduplication"
	"github.com/merbinr/deduplicator/internal/queue"
	"github.com/merbinr/deduplicator/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

var incoming_queue_conn queue.Queue_model

func CreateQueueClient() error {
	var err error
	username := config.Config.Services.IncommingQueue.User

	HOST_ENV := fmt.Sprintf("%s_DEDUPLICATOR_INCOMING_QUEUE_HOST", config.Config.StageName)

	host := os.Getenv(HOST_ENV)
	if host == "" {
		return fmt.Errorf("%s env is empty, please set it", HOST_ENV)
	}

	port := config.Config.Services.IncommingQueue.Port

	PASSWORD_ENV := fmt.Sprintf("%s_DEDUPLICATOR_INCOMING_QUEUE_PASSWORD", config.Config.StageName)
	password := os.Getenv(PASSWORD_ENV)
	if password == "" {
		return fmt.Errorf("%s env is empty, please set it", PASSWORD_ENV)
	}

	conn_string := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

	incoming_queue_conn.Conn, err = amqp.Dial(conn_string)
	if err != nil {
		return fmt.Errorf("unable to connect incoming queue, err: %s", err)
	}

	incoming_queue_conn.Channel, err = incoming_queue_conn.Conn.Channel()
	if err != nil {
		return fmt.Errorf("unable to create channel from connection in incomming queue, err: %s", err)
	}

	incoming_queue_conn.Queue, err = incoming_queue_conn.Channel.QueueDeclare(
		config.Config.Services.IncommingQueue.QueueName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("unable to create queue in incomming queue channel, err: %s", err)
	}
	return nil
}

func ConsumeMessage() {
	logger := logger.GetLogger()
	logger.Info("Trying to consume message from incoming queue")
	msgs, err := incoming_queue_conn.Channel.Consume(
		incoming_queue_conn.Queue.Name, // queue
		"",                             // consumer tag
		false,                          // auto-acknowledge
		false,                          // exclusive
		false,                          // no-local
		false,                          // no-wait
		nil,                            // arguments
	)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to consume message from incoming queue, err: %s", err))
	}

	for {
		select {
		case msg := <-msgs:
			if len(msg.Body) > 0 {
				logger.Debug(fmt.Sprintf("Sending message body into queue: %s", string(msg.Body)))
				err = deduplication.ProcessDeduplication(&msg.Body)
				if err != nil {
					logger.Error(fmt.Sprintf("unable to process deduplication, err: %s", err))
				}
				err = msg.Ack(true)
				if err != nil {
					logger.Error(fmt.Sprintf("unable to acknowledge the message, err: %s", err))
				}
			} else {
				logger.Error("message body is empty")
				time.Sleep(1 * time.Second)
			}
		default:
			logger.Debug("No message in queue, waiting for 1 second")
			time.Sleep(1 * time.Second)
		}
	}
}
