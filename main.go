package main

import (
	"flag"
	"log"
	"net/http"

	loggerPkg "github.com/vctaragao/sqs/pkg/logger"
	"github.com/vctaragao/sqs/pkg/queue"
	serverPkg "github.com/vctaragao/sqs/pkg/server"
	"github.com/vctaragao/sqs/pkg/sqs"
)

func main() {
	logger, err := loggerPkg.NewLogger("development.log")
	if err != nil {
		log.Fatal("creating logger", err)
	}

	port := flag.String("port", "7777", "specify the value of the port for the SQS to run")
	flag.Parse()

	sqsSvc := sqs.NewSQSService(queue.NewMessageQueue())

	serverMux := http.NewServeMux()
	sqsSvc.RegisterHandlers(serverMux)

	server := serverPkg.NewServer(logger)
	server.ListenAndServe(*port, serverMux)
}
