package incoming

import (
	"fmt"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var incomming_queue_conn queue.Queue_model

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

	incomming_queue_conn.Conn, err = amqp.Dial(conn_string)
	if err != nil {
		return fmt.Errorf("unable to connect incoming queue, err: %s", err)
	}

	incomming_queue_conn.Channel, err = incomming_queue_conn.Conn.Channel()
	if err != nil {
		return fmt.Errorf("unable to create channel from connection in incomming queue, err: %s", err)
	}

	incomming_queue_conn.Queue, err = incomming_queue_conn.Channel.QueueDeclare(
		config.Config.IncommingQueue.Name, // name
		false,                             // durable
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

func ConsumeMessage() ([]amqp.Delivery, error) {
	msgs, err := incomming_queue_conn.Channel.Consume(
		incomming_queue_conn.Queue.Name, // Queue name
		"",                              // Consumer tag
		false,                           // Auto-ack
		false,                           // Exclusive
		false,                           // No-local
		false,                           // No-wait
		nil,                             // Arguments
	)

	if err != nil {
		return []amqp.Delivery{}, fmt.Errorf("unable to consume message from incomming ")
	}

	messages := []amqp.Delivery{}
	current_number_of_msg := 1
	for msg := range msgs {
		messages = append(messages, msg)
		if current_number_of_msg >= 50 {
			return messages, nil
		}
		current_number_of_msg = current_number_of_msg + 1
	}
	return messages, nil
}
