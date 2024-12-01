package queue

import amqp "github.com/rabbitmq/amqp091-go"

type Queue_model struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}
