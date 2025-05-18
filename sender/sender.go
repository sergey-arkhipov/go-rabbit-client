// Package sender provides functionality for publishing messages to a RabbitMQ queue.
// It offers two approaches: a single-message publisher for simple tasks and a
// multi-goroutine publisher for high-throughput message sending.
package sender

import (
	"context"
	"encoding/json"
	"go-rabbit-client/common"
	"log"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Message represents a JSON-serializable message.
type Message struct {
	Content string `json:"content"`
}

// SendMessage publishes a single JSON message to a RabbitMQ queue.
// It is suitable for simple or low-volume publishing tasks.
func SendMessage(ch *amqp.Channel, config common.Config, message Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		config.ExchangeName, // exchange
		config.QueueName,    // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}
	log.Printf(" [x] Sent %s", string(body))
	return nil
}

// SendMessagesByGo publishes multiple JSON messages to a RabbitMQ queue using multiple goroutines.
// It is designed for high-throughput scenarios, distributing messages across numSenders goroutines.
// If numSenders is less than 1, it defaults to 1. The function returns an error if any message fails to send.
func SendMessagesByGo(conn *amqp.Connection, config common.Config, messages []Message, numSenders int) error {
	if len(messages) == 0 {
		return nil // No messages to send
	}

	// Validate numSenders
	if numSenders < 1 {
		numSenders = 1
	}

	// Calculate messages per sender
	msgsPerSender := (len(messages) + numSenders - 1) / numSenders
	var wg sync.WaitGroup
	errs := make(chan error, numSenders)

	// Start sender goroutines
	for i := range numSenders {
		startIdx := i * msgsPerSender
		endIdx := startIdx + msgsPerSender
		endIdx = int(math.Min(float64(endIdx), float64(len(messages))))

		if startIdx >= len(messages) {
			break // No more messages to process
		}

		senderMessages := messages[startIdx:endIdx]
		wg.Add(1)

		go func(senderID int, msgs []Message) {
			defer wg.Done()

			// Create a new channel for this goroutine
			ch, err := conn.Channel()
			if err != nil {
				errs <- err
				return
			}
			defer ch.Close()

			// Send messages
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			for _, msg := range msgs {
				body, err := json.Marshal(msg)
				if err != nil {
					errs <- err
					return
				}

				err = ch.PublishWithContext(ctx,
					config.ExchangeName, // exchange
					config.QueueName,    // routing key
					false,               // mandatory
					false,               // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					})
				if err != nil {
					errs <- err
					return
				}
				log.Printf(" [x] Sender %d sent %s", senderID, string(body))
			}
		}(i, senderMessages)
	}

	// Wait for all senders to complete
	go func() {
		wg.Wait()
		close(errs)
	}()

	// Collect errors
	for err := range errs {
		if err != nil {
			return err // Return first error encountered
		}
	}

	return nil
}
