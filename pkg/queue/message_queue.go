package queue

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmptyQueue   = errors.New("empty queue")
	ErrNotProcessed = errors.New("message has not yet been processed")
)

type Status int

const (
	Ready Status = iota
	Processing
	Finished
)

func (s Status) String() string {
	return []string{"ready", "processing", "finished"}[s]
}

type Message struct {
	ID      uuid.UUID       `json:"id"`
	Status  Status          `json:"status"`
	Payload json.RawMessage `json:"payload"`
}

func NewMessage(payload json.RawMessage) *Message {
	return &Message{
		ID:      uuid.New(),
		Status:  Ready,
		Payload: payload,
	}
}

func (m *Message) hasStatus(status Status) bool {
	return m.Status == status
}

type MessageQueue struct {
	Queue []*Message
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		Queue: make([]*Message, 0, 1000),
	}
}

func (q *MessageQueue) Enqueue(message *Message) {
	q.Queue = append(q.Queue, message)
}

func (q *MessageQueue) Read() *Message {
	if len(q.Queue) == 0 {
		return nil
	}

	message := q.Queue[0]
	if message.hasStatus(Processing) {
		return nil
	}

	message.Status = Processing

	return message
}

func (q *MessageQueue) Dequeue() (*Message, error) {
	if len(q.Queue) == 0 {
		return nil, ErrEmptyQueue
	}

	message := q.Queue[0]
	if message.hasStatus(Ready) {
		return nil, ErrNotProcessed
	}

	message.Status = Finished

	q.Queue = q.Queue[1:]

	return message, nil
}
