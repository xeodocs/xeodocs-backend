package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

// Project represents a project from the project service
type Project struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	RepoURL  string   `json:"repo_url"`
	Languages []string `json:"languages"`
}

// StartScheduler initializes and starts the cron scheduler
func StartScheduler(cfg *config.Config) {
	log.Println("Initializing Scheduler Service...")

	// Connect to RabbitMQ
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

	// Declare queue for sync_repo
	_, err = ch.QueueDeclare(
		"sync_repo", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue sync_repo: %v", err)
	}

	// Start cron scheduler
	c := cron.New()

	// Schedule sync repo job every hour
	_, err = c.AddFunc("@hourly", func() {
		syncAllRepos(cfg, ch)
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	c.Start()
	log.Println("Scheduler Service initialized successfully. Cron jobs started.")

	// Keep running
	select {}
}

// syncAllRepos fetches all projects and publishes sync_repo tasks
func syncAllRepos(cfg *config.Config, ch *amqp091.Channel) {
	log.Println("Running scheduled sync for all repos")

	// Fetch all projects from project service
	projects, err := fetchProjects(cfg)
	if err != nil {
		log.Printf("Failed to fetch projects: %v", err)
		return
	}

	for _, project := range projects.Projects {
		// Publish sync_repo task
		task := map[string]interface{}{
			"type": "sync_repo",
			"payload": map[string]interface{}{
				"projectId": project.ID,
			},
			"id": fmt.Sprintf("sync-%d-%d", project.ID, time.Now().Unix()),
		}

		body, err := json.Marshal(task)
		if err != nil {
			log.Printf("Failed to marshal task: %v", err)
			continue
		}

		err = ch.Publish(
			"",         // exchange
			"sync_repo", // routing key
			false,      // mandatory
			false,      // immediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			log.Printf("Failed to publish task for project %d: %v", project.ID, err)
		} else {
			log.Printf("Published sync_repo task for project %d", project.ID)
		}
	}

	// Log the cron job execution
	message := "Scheduled sync_repo tasks for all projects"
	logging.LogActivity(cfg.LoggingServiceURL, "cron_sync_repos", message, nil, nil, "cron")
}

// fetchProjects calls the project service to get all projects
func fetchProjects(cfg *config.Config) (*ProjectsResponse, error) {
	url := cfg.ProjectServiceURL + "/projects"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("project service returned status %d", resp.StatusCode)
	}

	var projectsResponse ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&projectsResponse); err != nil {
		return nil, err
	}

	return &projectsResponse, nil
}

// ProjectsResponse represents the response from /projects
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}
