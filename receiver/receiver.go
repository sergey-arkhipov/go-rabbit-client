// Package receiver provides functionality for consuming messages from a RabbitMQ queue.
// It offers two approaches: a single-consumer model for simple message processing and
// a multi-consumer model using goroutines for higher throughput.
package receiver

import (
	"go-rabbit-client/common"
	"log"
	"sync"

	"github.com/fatih/color"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ReceiveMessages consumes messages from a RabbitMQ queue using a single consumer.
// It uses automatic acknowledgments and is suitable for low-volume or simple processing tasks.
// The function blocks indefinitely until interrupted (e.g., with CTRL+C).
func ReceiveMessages(ch *amqp.Channel, config common.Config) error {
	msgs, err := ch.Consume(
		config.QueueName, // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message with routing key %s: %s", d.RoutingKey, string(d.Body))
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {} // Block indefinitely
}

// ReceiveMessagesByGo consumes messages from a RabbitMQ queue using multiple goroutines.
// It uses manual acknowledgments and is designed for higher throughput with configurable
// consumer counts. The numConsumers parameter specifies the number of goroutines; if less
// than 1, it defaults to 1. The function blocks indefinitely until interrupted (e.g., with CTRL+C).
func ReceiveMessagesByGo(ch *amqp.Channel, config common.Config, numConsumers int) error {
	// Prefetch limit per consumer
	const prefetchCount = 10

	// Validate numConsumers
	if numConsumers < 1 {
		numConsumers = 1 // Default to 1 if invalid
	}

	// Set QoS to limit unacknowledged messages per consumer
	err := ch.Qos(
		prefetchCount, // prefetch count
		0,             // prefetch size (0 = no limit)
		false,         // global (false = per consumer)
	)
	if err != nil {
		return err
	}

	// Wait group to ensure all consumers start before blocking
	var wg sync.WaitGroup
	wg.Add(numConsumers)

	// Start multiple consumer goroutines
	for i := range numConsumers {
		consumerID := i
		go func() {
			defer wg.Done()

			// Register consumer
			msgs, err := ch.Consume(
				config.QueueName, // queue
				"",               // consumer tag (auto-generated)
				false,            // auto-ack (manual acknowledgment)
				false,            // exclusive
				false,            // no-local
				false,            // no-wait
				nil,              // args
			)
			if err != nil {
				log.Printf("Consumer %d failed to register: %v", consumerID, err)
				return
			}

			// Process messages
			for d := range msgs {
				log.Printf("Consumer %d received a message with routing key %s: %s", consumerID, d.RoutingKey, string(d.Body))
				// Simulate processing (replace with actual processing logic)
				// Example: time.Sleep(100 * time.Millisecond)

				// Acknowledge the message
				if err := d.Ack(false); err != nil {
					log.Printf("Consumer %d failed to acknowledge message: %v", consumerID, err)
				}
			}
		}()
	}

	// Wait for all consumers to start
	color.Green(" [*] %d consumers waiting for messages. To exit press CTRL+C", numConsumers)
	wg.Wait()
	select {} // Block indefinitely
}
