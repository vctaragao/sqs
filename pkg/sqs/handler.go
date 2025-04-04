package sqs

import (
	"encoding/json"
	"net/http"

	"github.com/vctaragao/sqs/pkg/queue"
)

type SQS struct {
	messageQueue *queue.MessageQueue
}

func NewSQSService(messageQueue *queue.MessageQueue) *SQS {
	return &SQS{
		messageQueue: messageQueue,
	}
}

func (s *SQS) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.readMessage)
	mux.HandleFunc("POST /", s.queueMessage)
	mux.HandleFunc("DELETE /", s.removeMessage)
}

func (s *SQS) queueMessage(w http.ResponseWriter, r *http.Request) {
	var messageRequest json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&messageRequest); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "unable to decode message request body"}`))
		return
	}

	message := queue.NewMessage(messageRequest)
	s.messageQueue.Enqueue(message)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (s *SQS) readMessage(w http.ResponseWriter, r *http.Request) {
	message := s.messageQueue.Read()

	statusCode := http.StatusOK
	if message == nil {
		statusCode = http.StatusNoContent
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(message)
}

func (s *SQS) removeMessage(w http.ResponseWriter, r *http.Request) {
	message, err := s.messageQueue.Dequeue()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}
