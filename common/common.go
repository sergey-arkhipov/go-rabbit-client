// Package common provides utilities for connecting to a RabbitMQ server and managing queue configurations.
package common

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

// Config holds RabbitMQ connection and queue settings.
type Config struct {
	RabbitURL       string `yaml:"rabbit_url"`
	ExchangeName    string `yaml:"exchange_name"`
	QueueName       string `yaml:"queue_name"`
	QueueDurable    bool   `yaml:"queue_durable"`
	QueueAutoDelete bool   `yaml:"queue_auto_delete"`
	QueueExclusive  bool   `yaml:"queue_exclusive"`
	QueueNoWait     bool   `yaml:"queue_no_wait"`
}

// LoadConfig reads and parses the configuration from a YAML file.
func LoadConfig(filePath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

// Connect establishes a connection and channel to RabbitMQ using the provided configuration.
func Connect(config Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

// DeclareQueue declares a queue using the provided configuration.
func DeclareQueue(ch *amqp.Channel, config Config) error {
	_, err := ch.QueueDeclare(
		config.QueueName,       // name
		config.QueueDurable,    // durable
		config.QueueAutoDelete, // delete when unused
		config.QueueExclusive,  // exclusive
		config.QueueNoWait,     // no-wait
		nil,                    // arguments
	)
	return err
}
