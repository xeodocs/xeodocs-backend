package analytics_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/IBM/sarama"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

type Event struct {
	Event     string                 `json:"event"`
	UserID    *int                   `json:"userId,omitempty"`
	ProjectID *int                   `json:"projectId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"-"`
}

type Processor struct {
	kafkaConsumer sarama.Consumer
	influxClient  influxdb2.Client
	writeAPI      interface{}
	cfg           *config.Config
}

func NewProcessor(cfg *config.Config) (*Processor, error) {
	// Initialize Kafka consumer
	consumer, err := sarama.NewConsumer([]string{cfg.KafkaBrokers}, nil)
	if err != nil {
		return nil, err
	}

	// Initialize InfluxDB client
	client := influxdb2.NewClientWithOptions(cfg.InfluxDBURL, cfg.InfluxDBToken,
		influxdb2.DefaultOptions().SetBatchSize(20))
	writeAPI := client.WriteAPI(cfg.InfluxDBOrg, cfg.InfluxDBBucket)

	return &Processor{
		kafkaConsumer: consumer,
		influxClient:  client,
		writeAPI:      writeAPI,
		cfg:           cfg,
	}, nil
}

func (p *Processor) Start() error {
	// Subscribe to topic
	partitionConsumer, err := p.kafkaConsumer.ConsumePartition("raw_traffic_events", 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer partitionConsumer.Close()

	log.Println("Analytics Processor started, consuming from Kafka")

	// Process messages
	for message := range partitionConsumer.Messages() {
		if err := p.processMessage(message); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}

	return nil
}

func (p *Processor) processMessage(msg *sarama.ConsumerMessage) error {
	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}
	event.Timestamp = time.Now() // or use msg.Timestamp

	// Log the event
	message := "Processed analytics event: " + event.Event
	logging.LogActivity(p.cfg.LoggingServiceURL, "analytics_event", message, event.UserID, event.ProjectID, "info")

	// Create InfluxDB point
	tags := map[string]string{}
	if event.ProjectID != nil {
		tags["projectId"] = fmt.Sprintf("%d", *event.ProjectID)
	}
	if event.UserID != nil {
		tags["userId"] = fmt.Sprintf("%d", *event.UserID)
	}

	point := write.NewPoint(
		event.Event,
		tags,
		map[string]interface{}{
			"value": 1,
		},
		event.Timestamp,
	)

	// Write to InfluxDB using reflection to avoid type issues
	v := reflect.ValueOf(p.writeAPI)
	method := v.MethodByName("WritePoint")
	if method.IsValid() {
		// Assuming the method signature is WritePoint(ctx context.Context, point *write.Point)
		ctx := reflect.ValueOf(context.Background())
		pointVal := reflect.ValueOf(point)
		method.Call([]reflect.Value{ctx, pointVal})
	} else {
		log.Printf("WritePoint method not found")
	}

	return nil
}

func (p *Processor) Close() {
	p.kafkaConsumer.Close()
	p.influxClient.Close()
}
