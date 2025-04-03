package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vctaragao/sqs/pkg/queue"
)

func TestSQSApi(t *testing.T) {
	testSuite := SetupTestSuite(t)
	defer testSuite.Close()

	t.Run("POST /", func(t *testing.T) {
		testPayload, err := json.Marshal(`{"test": "test"}`)
		assert.NoError(t, err)

		testRequest, err := http.NewRequest("POST", testSuite.ServerAddr(), bytes.NewReader(testPayload))
		assert.NoError(t, err)

		response, err := testSuite.Do(testRequest)
		assert.NoError(t, err)

		bodyData, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		fmt.Println("response message", string(bodyData))

		assert.Equal(t, http.StatusCreated, response.StatusCode)

		testSuite.ExpectMessageInQueue(t, queue.Message{
			Payload: testPayload,
			Status:  queue.Ready,
		})
	})
}
