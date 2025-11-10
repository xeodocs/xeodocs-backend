package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rabbitmq/amqp091-go"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
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
		go consumeQueue(cfg, ch, queue)
	}

	log.Println("Worker Service initialized successfully. Listening for messages...")
	select {} // Keep running
}

func consumeQueue(cfg *config.Config, ch *amqp091.Channel, queueName string) {
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
			processTask(cfg, task)

			d.Ack(false) // acknowledge the message
		}
	}()

	log.Printf("Started consuming queue: %s", queueName)
	<-forever
}

func processTask(cfg *config.Config, task Task) {
	log.Printf("Processing task type: %s with payload: %v", task.Type, task.Payload)

	// Log task processing start
	message := fmt.Sprintf("Worker processing task: %s", task.Type)
	logging.LogActivity(cfg.LoggingServiceURL, "task_processing", message, nil, nil, "info")

	switch task.Type {
	case "clone_repo":
		handleCloneRepo(cfg, task.Payload)
	case "create_language_copies":
		handleCreateLanguageCopies(cfg, task.Payload)
	case "sync_repo":
		handleSyncRepo(cfg, task.Payload)
	case "delete_repo":
		handleDeleteRepo(cfg, task.Payload)
	case "build_task":
		handleBuildTask(cfg, task.Payload)
	default:
		log.Printf("Unknown task type: %s", task.Type)
	}
}

func handleCloneRepo(cfg *config.Config, payload map[string]interface{}) {
	repoURL, ok1 := payload["repoUrl"].(string)
	projectIDFloat, ok2 := payload["projectId"].(float64)

	if !ok1 || !ok2 {
		log.Printf("Invalid payload for clone_repo: %v", payload)
		return
	}

	projectID := int(projectIDFloat)

	req := map[string]interface{}{
		"repoUrl":   repoURL,
		"projectId": projectID,
	}

	if err := callRepositoryService(cfg, http.MethodPost, "/internal/clone-repo", req); err != nil {
		log.Printf("Failed to clone repo: %v", err)
	} else {
		// Log successful repo cloning
		message := fmt.Sprintf("Worker successfully cloned repo for project %d", projectID)
		logging.LogActivity(cfg.LoggingServiceURL, "worker_repo_cloned", message, nil, &projectID, "info")
	}
}

func handleCreateLanguageCopies(cfg *config.Config, payload map[string]interface{}) {
	projectIDFloat, ok1 := payload["projectId"].(float64)
	languagesInterface, ok2 := payload["languages"].([]interface{})

	if !ok1 || !ok2 {
		log.Printf("Invalid payload for create_language_copies: %v", payload)
		return
	}

	projectID := int(projectIDFloat)
	languages := make([]string, len(languagesInterface))
	for i, lang := range languagesInterface {
		if langStr, ok := lang.(string); ok {
			languages[i] = langStr
		}
	}

	req := map[string]interface{}{
		"projectId": projectID,
		"languages": languages,
	}

	if err := callRepositoryService(cfg, http.MethodPost, "/internal/create-language-copies", req); err != nil {
		log.Printf("Failed to create language copies: %v", err)
	} else {
		// Log successful language copies
		message := fmt.Sprintf("Worker successfully created language copies for project %d", projectID)
		logging.LogActivity(cfg.LoggingServiceURL, "worker_language_copies_created", message, nil, &projectID, "info")
	}
}

func handleSyncRepo(cfg *config.Config, payload map[string]interface{}) {
	projectIDFloat, ok := payload["projectId"].(float64)

	if !ok {
		log.Printf("Invalid payload for sync_repo: %v", payload)
		return
	}

	projectID := int(projectIDFloat)

	req := map[string]interface{}{
		"projectId": projectID,
	}

	if err := callRepositoryService(cfg, http.MethodPut, "/internal/sync-repo", req); err != nil {
		log.Printf("Failed to sync repo: %v", err)
	} else {
		// Log successful repo sync
		message := fmt.Sprintf("Worker successfully synced repo for project %d", projectID)
		logging.LogActivity(cfg.LoggingServiceURL, "worker_repo_synced", message, nil, &projectID, "info")
	}
}

func handleDeleteRepo(cfg *config.Config, payload map[string]interface{}) {
	projectIDFloat, ok := payload["projectId"].(float64)

	if !ok {
		log.Printf("Invalid payload for delete_repo: %v", payload)
		return
	}

	projectID := int(projectIDFloat)

	req := map[string]interface{}{
		"projectId": projectID,
	}

	if err := callRepositoryService(cfg, http.MethodDelete, "/internal/delete-repo", req); err != nil {
		log.Printf("Failed to delete repo: %v", err)
	} else {
		// Log successful repo deletion
		message := fmt.Sprintf("Worker successfully deleted repo for project %d", projectID)
		logging.LogActivity(cfg.LoggingServiceURL, "worker_repo_deleted", message, nil, &projectID, "info")
	}
}

func handleBuildTask(cfg *config.Config, payload map[string]interface{}) {
	projectIDFloat, ok1 := payload["projectId"].(float64)
	buildType, ok2 := payload["buildType"].(string) // "build", "export", or "preview"

	if !ok1 || !ok2 {
		log.Printf("Invalid payload for build_task: %v", payload)
		return
	}

	projectID := int(projectIDFloat)

	req := map[string]interface{}{
		"projectId": projectID,
	}

	var endpoint string
	switch buildType {
	case "build":
		endpoint = "/internal/build"
	case "export":
		endpoint = "/internal/export"
	case "preview":
		endpoint = "/internal/preview"
	default:
		log.Printf("Unknown build type: %s", buildType)
		return
	}

	if err := callBuildService(cfg, http.MethodPost, endpoint, req); err != nil {
		log.Printf("Failed to execute %s: %v", buildType, err)
	} else {
		// Log successful build task
		message := fmt.Sprintf("Worker successfully executed %s for project %d", buildType, projectID)
		logging.LogActivity(cfg.LoggingServiceURL, "worker_build_task_completed", message, nil, &projectID, "info")
	}
}

func callRepositoryService(cfg *config.Config, method, endpoint string, req map[string]interface{}) error {
	url := cfg.RepositoryServiceURL + endpoint

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var resp *http.Response
	switch method {
	case http.MethodPost:
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	case http.MethodPut:
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
	case http.MethodDelete:
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("repository service returned status %d", resp.StatusCode)
	}

	log.Printf("Successfully called %s %s", method, endpoint)
	return nil
}

func callBuildService(cfg *config.Config, method, endpoint string, req map[string]interface{}) error {
	url := cfg.BuildServiceURL + endpoint

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("build service returned status %d", resp.StatusCode)
	}

	log.Printf("Successfully called %s %s", method, endpoint)
	return nil
}
