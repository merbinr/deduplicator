package outgoing

import (
	"fmt"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var outgoing_queue_conn queue.Queue_model

func CreateQueueClient() error {
	var err error
	username := config.Config.QutgoingQueue.User
	port := config.Config.QutgoingQueue.Port
	password := os.Getenv("DEDUPLICATOR_OUTGOING_QUEUE_PASSWORD")
	if password == "" {
		return fmt.Errorf("DEDUPLICATOR_OUTGOING_QUEUE_PASSWORD env is empty, please set it")
	}

	host := os.Getenv("DEDUPLICATOR_OUTGOING_QUEUE_HOST")
	if host == "" {
		return fmt.Errorf("DEDUPLICATOR_OUTGOING_QUEUE_HOST env is empty, please set it")
	}

	conn_string := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	outgoing_queue_conn.Conn, err = amqp.Dial(conn_string)
	if err != nil {
		return fmt.Errorf("unable to connect incoming queue, err: %s", err)
	}

	outgoing_queue_conn.Channel, err = outgoing_queue_conn.Conn.Channel()
	if err != nil {
		return fmt.Errorf("unable to create channel from connection in ouitgoing queue, err: %s", err)
	}

	outgoing_queue_conn.Queue, err = outgoing_queue_conn.Channel.QueueDeclare(
		config.Config.QutgoingQueue.Name, // name
		true,                             // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		return fmt.Errorf("unable to create queue in outgoing queue channel, err: %s", err)
	}
	return nil
}

func SendMessage(message []byte) error {
	err := outgoing_queue_conn.Channel.Publish(
		"",                             // exchange
		outgoing_queue_conn.Queue.Name, // routing key (queue name)
		false,                          // mandatory
		false,                          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to publish message to the queue, err: %s", err)
	}
	return nil
}
