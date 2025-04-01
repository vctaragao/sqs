package sqs

import (
	"encoding/json"
	"net/http"

	"github.com/vctaragao/sqs/pkg/queue"
)

type SQS struct {
	messageQueue queue.MessageQueue
}

func NewSQSService() *SQS {
	return &SQS{
		messageQueue: queue.NewMessageQueue(),
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

	s.messageQueue.Queue(queue.NewMessage(messageRequest))
}

func (s *SQS) readMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.messageQueue.Read())
}

func (s *SQS) removeMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.messageQueue.Dequeue())
}
