// Package main
package main

import (
	"go-rabbit-client/common"
	"go-rabbit-client/receiver"
	"go-rabbit-client/sender"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Load configuration
	config, err := common.LoadConfig("config.yml")
	failOnError(err, "Failed to load config")

	// Establish connection and channel
	conn, ch, err := common.Connect(config)
	failOnError(err, "Failed to connect or open channel")

	defer common.CloseWithLog(conn, "failed to close Rabbit connection")
	defer common.CloseWithLog(ch, "failed to close Rabbit channel")

	// Attempt to declare queue with config settings
	err = common.DeclareQueue(ch, config)
	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == 406 { // PRECONDITION_FAILED
			if strings.Contains(amqpErr.Reason, "inequivalent arg 'durable'") {
				log.Fatalf("Queue '%s' exists with durable=false, but config requires durable=%t. Please delete the queue or update its settings.", config.QueueName, config.QueueDurable)
			}
		}
		failOnError(err, "Failed to declare queue")
	}

	// Demonstrate sending multiple JSON messages with goroutines
	messages := []sender.Message{
		{Content: "Hello World! 1"},
		{Content: "Hello World! 2"},
		{Content: "Hello World! 3"},
		{Content: "Hello World! 4"},
	}
	err = sender.SendMessagesByGo(conn, config, messages, 2) // Use 2 senders
	failOnError(err, "Failed to send messages")

	// Wait briefly to ensure messages are processed
	time.Sleep(500 * time.Millisecond)

	// Demonstrate receiving messages with multiple consumers
	err = receiver.ReceiveMessagesByGo(ch, config, 4) // Use 4 consumers
	failOnError(err, "Failed to receive messages")
}
