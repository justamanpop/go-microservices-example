package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "8080"

type Config struct {
	RabbitConn *amqp.Connection
}

func main() {
	log.Println("Starting broker service on port", webPort)

	conn, err := connect()
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	app := Config{
		RabbitConn: conn,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = server.ListenAndServe()

	if err != nil {
		log.Panic(err)
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
