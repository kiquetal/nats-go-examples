// Package pubsub provides NATS publish/subscribe functionality
package pubsub

import (
	"encoding/json"
	"time"

	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/nats-io/nats.go"
)

// Publisher defines the interface for publishing messages
type Publisher interface {
	Publish(subject string, data []byte) error
	PublishMessage(msg *models.Message) error
	Close()
}

// NATSPublisher implements the Publisher interface using NATS
type NATSPublisher struct {
	conn *nats.Conn
}

// NewPublisher creates a new NATS publisher
func NewPublisher(natsURL string, options ...nats.Option) (*NATSPublisher, error) {
	// Set default connection timeout
	opts := append([]nats.Option{nats.Timeout(10 * time.Second)}, options...)

	// Connect to NATS
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, err
	}

	return &NATSPublisher{conn: nc}, nil
}

// Publish sends a raw byte message to the specified subject
func (p *NATSPublisher) Publish(subject string, data []byte) error {
	return p.conn.Publish(subject, data)
}

// PublishMessage serializes and publishes a Message
func (p *NATSPublisher) PublishMessage(msg *models.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.Publish(msg.Subject, data)
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}
