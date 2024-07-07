package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	queue, err := declareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := channel.QueueBind(queue.Name, s, "logs_topic", false, nil)
		if err != nil {
			return err
		}
	}
	messages, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	myChannel := make(chan bool)

	go func() {
		for message := range messages {
			var payload Payload
			_ = json.Unmarshal(message.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Listening for messages on [Exchange, Queue], [logs_topic, %s]\n", queue.Name)
	<-myChannel //to make it run forever
	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log":
		err := logEvent(payload)
		if err != nil {
			log.Println("something went wrong in logging event from listener", err)
		}
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println("something went wrong in logging event from listener", err)
		}
	}
}

func logEvent(payload Payload) error {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")

	//docker compose takes care of resolving the URL here
	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}
