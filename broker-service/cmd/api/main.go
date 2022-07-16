package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const webPort = "80"

type Config struct {
	MaxFileSize int
	Rabbit      *amqp.Connection
}

func main() {
	// try to connect rabbitmq
	rabbitConn, err := connect()
	failOnError(err, "Failed to connect to RabbitMQ")
	defer rabbitConn.Close()

	app := Config{
		MaxFileSize: 0,
		Rabbit:      rabbitConn,
	}

	log.Printf("Staring boroker services at %s", webPort)
	// define server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	// start server
	err = srv.ListenAndServe()

	if err != nil {
		log.Panicf("Err run server")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
		os.Exit(1)
	}
}

func connect() (*amqp.Connection, error) {
	counts := 0
	var backOff = 1 * time.Second
	var connection *amqp.Connection
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
		if err != nil {
			fmt.Printf("RabiitMQ not yet ready")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off")
		time.Sleep(backOff)
		continue
	}

	return connection, nil

}
