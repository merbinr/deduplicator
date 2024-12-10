package incoming

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/deduplication"
	"github.com/merbinr/deduplicator/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var incoming_queue_conn queue.Queue_model

func CreateQueueClient() error {
	var err error
	username := config.Config.IncommingQueue.User

	host := os.Getenv("DEDUPLICATOR_INCOMING_QUEUE_HOST")
	if host == "" {
		return fmt.Errorf("DEDUPLICATOR_INCOMING_QUEUE_HOST env is empty, please set it")
	}

	port := config.Config.IncommingQueue.Port
	password := os.Getenv("DEDUPLICATOR_INCOMING_QUEUE_PASSWORD")
	if password == "" {
		return fmt.Errorf("DEDUPLICATOR_INCOMING_QUEUE_PASSWORD env is empty, please set it")
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
		config.Config.IncommingQueue.Name, // name
		true,                              // durable
		false,                             // delete when unused
		false,                             // exclusive
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		return fmt.Errorf("unable to create queue in incomming queue channel, err: %s", err)
	}
	return nil
}

func ConsumeMessage() error {
	slog.Info("Trying to consume message from incoming queue")
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
		return err
	}

	for msg := range msgs {
		err = deduplication.ProcessDeduplication(msg.Body)
		if err != nil {
			slog.Error(fmt.Sprintf("unable to process deduplication, err: %s", err))
		}
		err = msg.Ack(true)
		if err != nil {
			slog.Error(fmt.Sprintf("unable to acknowledge the message, err: %s", err))
		}
	}
	slog.Info("Channel closed, breaking the loop")
	return nil
}
