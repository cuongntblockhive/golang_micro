package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Emiter struct {
	connection *amqp.Connection
}

func (e *Emiter) Setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return declareExchange(channel)
}

func (e *Emiter) Push(event string, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	log.Println("Pushing to channel")

	err = channel.Publish("logs_topic", severity, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(event)})
	if err != nil {
		return err
	}
	return nil
}

func NewEventEmitter(conn *amqp.Connection) (Emiter, error) {
	emiter := Emiter{
		connection: conn,
	}
	err := emiter.Setup()
	if err != nil {
		return Emiter{}, err
	}
	return emiter, nil
}
