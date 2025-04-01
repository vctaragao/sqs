package queue

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Status int

const (
	Ready Status = iota
	Processing
	Finished
)

type Message struct {
	ID      uuid.UUID
	Status  Status
	Payload json.RawMessage
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
	queue []*Message
}

func NewMessageQueue() MessageQueue {
	return MessageQueue{
		queue: make([]*Message, 1000),
	}
}

func (q MessageQueue) Queue(message *Message) {
	q.queue = append(q.queue, message)
}

func (q MessageQueue) Read() *Message {
	if len(q.queue) == 0 {
		return nil
	}

	message := q.queue[0]
	if message.hasStatus(Processing) {
		return nil
	}

	message.Status = Processing

	return message
}

func (q MessageQueue) Dequeue() *Message {
	if len(q.queue) == 0 {
		return nil
	}

	message := q.queue[0]
	message.Status = Finished

	q.queue = q.queue[1:]

	return message
}
