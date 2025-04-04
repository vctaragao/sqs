package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vctaragao/sqs/pkg/queue"
	"github.com/vctaragao/sqs/pkg/sqs"
)

type TestSuite struct {
	server       *httptest.Server
	messageQueue *queue.MessageQueue
}

func SetupTestSuite(t *testing.T) TestSuite {
	t.Helper()

	messageQueue := queue.NewMessageQueue()
	sqsSvc := sqs.NewSQSService(messageQueue)

	testServerMux := http.NewServeMux()
	sqsSvc.RegisterHandlers(testServerMux)

	testServer := httptest.NewUnstartedServer(testServerMux)

	testSuite := TestSuite{
		server:       testServer,
		messageQueue: messageQueue,
	}

	testServer.Start()

	return testSuite
}

func (ts *TestSuite) Close() {
	ts.server.Close()
}

func (ts *TestSuite) Do(r *http.Request) (*http.Response, error) {
	return ts.server.Client().Do(r)
}
func (ts *TestSuite) ServerAddr() string {
	return "http://" + ts.server.Listener.Addr().String()
}

func (ts *TestSuite) ExpectMessageInQueue(t *testing.T, expectedMessage queue.Message) {
	t.Helper()

	message := ts.messageQueue.Queue[0]
	require.NotNil(t, message)

	assert.Equal(t, expectedMessage.ID, message.ID)
	assert.JSONEq(t, string(expectedMessage.Payload), string(message.Payload))
	assert.Equal(t, expectedMessage.Status, message.Status, fmt.Sprintf("has: %s, want: %s\n", message.Status, expectedMessage.Status))
}

func (ts *TestSuite) CountInQueue(t *testing.T, size int) {
	t.Helper()

	assert.Len(t, ts.messageQueue.Queue, size)
}

func (ts *TestSuite) ClenQueue() {
	*ts.messageQueue = *queue.NewMessageQueue()
}

func (ts *TestSuite) TestCleanup(t *testing.T, testFunc func(t *testing.T)) func(t *testing.T) {
	t.Helper()

	return func(t *testing.T) {
		t.Cleanup(ts.ClenQueue)
		testFunc(t)
	}
}

func (ts *TestSuite) SeedQueue(message *queue.Message) {
	ts.messageQueue.Queue = append(ts.messageQueue.Queue, message)
}
