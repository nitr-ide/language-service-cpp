package main

import "log"

func init() {

	log.Println("Starting CPP Service")

	err := connectQueue()

	if err != nil {
		log.Panicln(err)
	} else {
		log.Println("Successfully connected to RabbitMQ:5672")
	}

	log.Println("Successfully initalised")

}

func main() {

	defer func() {
		queueConf.channel.Close()
		queueConf.conn.Close()
		log.Println("Successfully closed RabbitMQ")
	}()

	startConsumer()

}
