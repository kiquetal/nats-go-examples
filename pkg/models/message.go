// Package models contains data structures shared across the application
package models

import "time"

// Message represents a generic message structure for NATS communication
type Message struct {
	ID        string            `json:"id"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewMessage creates a new Message with the given subject and body
func NewMessage(subject, body string) *Message {
	return &Message{
		ID:        generateID(),
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// AddMetadata adds a key-value pair to the message metadata
func (m *Message) AddMetadata(key, value string) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]string)
	}
	m.Metadata[key] = value
}

// Helper function to generate a simple unique ID
func generateID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

// Helper function to generate a random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
