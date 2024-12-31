package outgoing

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var outgoing_queue_conn queue.Queue_model

func CreateQueueClient() error {
	var err error
	username := config.Config.Services.OutgoingQueue.User
	port := config.Config.Services.OutgoingQueue.Port

	PASSWORD_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OUTGOING_QUEUE_PASSWORD", config.Config.StageName)
	password := os.Getenv(PASSWORD_ENV)
	if password == "" {
		return fmt.Errorf("%s env is empty, please set it", PASSWORD_ENV)
	}

	HOST_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OUTGOING_QUEUE_HOST", config.Config.StageName)
	host := os.Getenv(HOST_ENV)
	if host == "" {
		return fmt.Errorf("%s env is empty, please set it", HOST_ENV)
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
		config.Config.Services.OutgoingQueue.QueueName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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
	slog.Debug("Successfully send message to the outgoing queue")
	return nil
}
