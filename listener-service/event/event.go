package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(
		"logs_topic", //exchange name
		"topic",      //type
		true,         //durable?
		false,        //auto-deleted after use?
		false,        //internal?
		false,        //no-wait?
		nil,          //arguments
	)
}

func declareRandomQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    //name
		false, //delete after use?
		false, //durable?
		true,  //exclusive to channel?
		false, //no-wait?
		nil,   //arguments
	)
}
