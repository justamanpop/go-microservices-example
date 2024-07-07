package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/justamanpop/microservices_listener/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	//connect to rabbitmq
	conn, err := connect()
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	consumer, err := event.NewConsumer(conn)
	if err != nil {
		log.Panic(err)
	}

	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var attempts int
	backoff := 1 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not ready yet...")
			attempts += 1
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = c
			break
		}

		if attempts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(attempts), 2)) * time.Second
		log.Println("Backing off...")
		time.Sleep(backoff)
		continue
	}
	return connection, nil
}
