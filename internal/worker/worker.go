package worker

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

// Start initializes the worker service and begins consuming messages from RabbitMQ
func Start(cfg *config.Config) {
	log.Println("Initializing Worker Service...")

	conn, err := amqp091.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare queues
	queues := []string{"clone_repo", "translate_files", "build_task"}
	for _, queue := range queues {
		_, err := ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", queue, err)
		}
		log.Printf("Declared queue: %s", queue)
	}

	// Start consuming from each queue
	for _, queue := range queues {
		go consumeQueue(ch, queue)
	}

	log.Println("Worker Service initialized successfully. Listening for messages...")
	select {} // Keep running
}

func consumeQueue(ch *amqp091.Channel, queueName string) {
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer for queue %s: %v", queueName, err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var task Task
			if err := json.Unmarshal(d.Body, &task); err != nil {
				log.Printf("Failed to unmarshal task from queue %s: %v", queueName, err)
				d.Nack(false, false) // negative acknowledge, don't requeue
				continue
			}

			log.Printf("Received task: %s, ID: %s", task.Type, task.ID)

			// Process the task (placeholder)
			processTask(task)

			d.Ack(false) // acknowledge the message
		}
	}()

	log.Printf("Started consuming queue: %s", queueName)
	<-forever
}

func processTask(task Task) {
	log.Printf("Processing task type: %s with payload: %v", task.Type, task.Payload)
	// TODO: Implement actual task processing logic
	// e.g., call Repository Service for clone_repo, Translation Service for translate_files, etc.
}
