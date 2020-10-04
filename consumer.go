package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

var queueConf struct {
	url     string
	conn    *amqp.Connection
	channel *amqp.Channel
}

var queueNames map[string]string

func connectQueue() error {

	queueConf.url = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(queueConf.url)

	if err != nil {
		return err
	}

	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	queueConf.channel = ch
	queueConf.conn = conn

	_, err = ch.QueueDeclare(
		"PROCESS_CPP",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	return nil

}

func startConsumer() error {

	delivery, err := queueConf.channel.Consume(
		"PROCESS_CPP",
		"",
		false,
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	log.Println("Starting CPP Consumer")

	for {
		select {
		case d := <-delivery:

			var message Request

			err := json.Unmarshal(d.Body, &message)

			if err != nil {
				log.Println("error while parsing", err)
				continue
			}

			log.Println("Recieved Job", message.Language)

			d.Ack(true)

			HandleCpp(&message)

		}
	}

}
