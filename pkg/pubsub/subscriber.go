// Package pubsub provides NATS publish/subscribe functionality
package pubsub

import (
	"encoding/json"
	"time"

	"github.com/kiquetal/nats-go-examples/pkg/models"
	"github.com/nats-io/nats.go"
)

// MessageHandler is a function type for handling received messages
type MessageHandler func(*models.Message) error

// RawMessageHandler is a function type for handling raw message data
type RawMessageHandler func(subject string, data []byte) error

// Subscriber defines the interface for subscribing to messages
type Subscriber interface {
	Subscribe(subject string, handler RawMessageHandler) (*nats.Subscription, error)
	SubscribeMessage(subject string, handler MessageHandler) (*nats.Subscription, error)
	QueueSubscribe(subject, queue string, handler RawMessageHandler) (*nats.Subscription, error)
	QueueSubscribeMessage(subject, queue string, handler MessageHandler) (*nats.Subscription, error)
	Close()
}

// NATSSubscriber implements the Subscriber interface using NATS
type NATSSubscriber struct {
	conn *nats.Conn
}

// NewSubscriber creates a new NATS subscriber
func NewSubscriber(natsURL string, options ...nats.Option) (*NATSSubscriber, error) {
	// Set default connection timeout
	opts := append([]nats.Option{nats.Timeout(10 * time.Second)}, options...)

	// Connect to NATS
	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, err
	}

	return &NATSSubscriber{conn: nc}, nil
}

// Subscribe subscribes to a subject with a raw message handler
func (s *NATSSubscriber) Subscribe(subject string, handler RawMessageHandler) (*nats.Subscription, error) {
	return s.conn.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(msg.Subject, msg.Data); err != nil {
			// Handle error (could log here)
		}
	})
}

// SubscribeMessage subscribes to a subject with a structured message handler
func (s *NATSSubscriber) SubscribeMessage(subject string, handler MessageHandler) (*nats.Subscription, error) {
	return s.conn.Subscribe(subject, func(msg *nats.Msg) {
		var message models.Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			// Handle error (could log here)
			return
		}

		if err := handler(&message); err != nil {
			// Handle error (could log here)
		}
	})
}

// QueueSubscribe subscribes to a subject with a queue group and raw message handler
func (s *NATSSubscriber) QueueSubscribe(subject, queue string, handler RawMessageHandler) (*nats.Subscription, error) {
	return s.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler(msg.Subject, msg.Data); err != nil {
			// Handle error (could log here)
		}
	})
}

// QueueSubscribeMessage subscribes to a subject with a queue group and structured message handler
func (s *NATSSubscriber) QueueSubscribeMessage(subject, queue string, handler MessageHandler) (*nats.Subscription, error) {
	return s.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var message models.Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			// Handle error (could log here)
			return
		}

		if err := handler(&message); err != nil {
			// Handle error (could log here)
		}
	})
}

// Close closes the NATS connection
func (s *NATSSubscriber) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}
