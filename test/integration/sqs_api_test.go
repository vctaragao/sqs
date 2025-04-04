package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vctaragao/sqs/pkg/queue"
)

func TestSQSApi(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Close()

	testPayload, err := json.Marshal(`{"test": "test"}`)
	assert.NoError(t, err)

	t.Run("POST /", ts.TestCleanup(t, func(t *testing.T) {
		testRequest, err := http.NewRequest("POST", ts.ServerAddr(), bytes.NewReader(testPayload))
		assert.NoError(t, err)

		response, err := ts.Do(testRequest)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, response.StatusCode)

		var createdMessage queue.Message
		err = json.NewDecoder(response.Body).Decode(&createdMessage)
		assert.NoError(t, err)

		defer response.Body.Close()

		assert.JSONEq(t, string(testPayload), string(createdMessage.Payload))
		assert.Equal(t, queue.Ready, createdMessage.Status)

		ts.ExpectMessageInQueue(t, queue.Message{
			ID:      createdMessage.ID,
			Payload: testPayload,
			Status:  queue.Ready,
		})
	}))

	t.Run("GET /", func(t *testing.T) {
		t.Run("should return 204 on empty queue", ts.TestCleanup(t, func(t *testing.T) {
			testRequest, err := http.NewRequest("GET", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusNoContent, response.StatusCode)

			ts.CountInQueue(t, 0)
		}))

		t.Run("should return 200 with a message", ts.TestCleanup(t, func(t *testing.T) {
			message := queue.NewMessage(testPayload)
			ts.SeedQueue(message)

			testRequest, err := http.NewRequest("GET", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, response.StatusCode)

			var messageRead queue.Message
			err = json.NewDecoder(response.Body).Decode(&messageRead)
			assert.NoError(t, err)

			defer response.Body.Close()

			assert.Equal(t, message.ID, messageRead.ID)
			assert.Equal(t, queue.Processing, messageRead.Status)
			assert.JSONEq(t, string(message.Payload), string(messageRead.Payload))

			ts.CountInQueue(t, 1)
			ts.ExpectMessageInQueue(t, queue.Message{
				ID:      messageRead.ID,
				Status:  messageRead.Status,
				Payload: messageRead.Payload,
			})
		}))

		t.Run("should return 204 with a message in processing", ts.TestCleanup(t, func(t *testing.T) {
			message := queue.NewMessage(testPayload)
			message.Status = queue.Processing
			ts.SeedQueue(message)

			testRequest, err := http.NewRequest("GET", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusNoContent, response.StatusCode)

			ts.CountInQueue(t, 1)
			ts.ExpectMessageInQueue(t, queue.Message{
				ID:      message.ID,
				Status:  message.Status,
				Payload: message.Payload,
			})
		}))
	})

	t.Run("DELETE /", func(t *testing.T) {
		t.Run("should remove finished message", ts.TestCleanup(t, func(t *testing.T) {
			message := queue.NewMessage(testPayload)
			message.Status = queue.Processing
			ts.SeedQueue(message)

			testRequest, err := http.NewRequest("DELETE", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, response.StatusCode)

			var messageDeQueued queue.Message
			err = json.NewDecoder(response.Body).Decode(&messageDeQueued)
			assert.NoError(t, err)

			defer response.Body.Close()

			assert.Equal(t, message.ID, messageDeQueued.ID)
			assert.Equal(t, queue.Finished, messageDeQueued.Status)
			assert.JSONEq(t, string(message.Payload), string(messageDeQueued.Payload))

			ts.CountInQueue(t, 0)
		}))

		t.Run("should return err if try to delete message without been processed", ts.TestCleanup(t, func(t *testing.T) {
			message := queue.NewMessage(testPayload)
			ts.SeedQueue(message)

			testRequest, err := http.NewRequest("DELETE", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			var responseErr string
			err = json.NewDecoder(response.Body).Decode(&responseErr)
			assert.NoError(t, err)

			defer response.Body.Close()

			assert.Equal(t, queue.ErrNotProcessed.Error(), responseErr)

			ts.CountInQueue(t, 1)
		}))

		t.Run("should return err if try to delete with an empty queue", ts.TestCleanup(t, func(t *testing.T) {
			testRequest, err := http.NewRequest("DELETE", ts.ServerAddr(), nil)
			assert.NoError(t, err)

			response, err := ts.Do(testRequest)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			var responseErr string
			err = json.NewDecoder(response.Body).Decode(&responseErr)
			assert.NoError(t, err)

			fmt.Println(responseErr)

			defer response.Body.Close()

			assert.Equal(t, queue.ErrEmptyQueue.Error(), responseErr)

			ts.CountInQueue(t, 0)
		}))
	})
}
