package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	logger *log.Logger
}

func NewServer(logger *log.Logger) Server {
	return Server{
		logger: logger,
	}
}

func (s *Server) ListenAndServe(port string, handler *http.ServeMux) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("ok")); err != nil {
			s.logger.Printf("ERR: Writing response: %v", err)
		}
	})

	go func() {
		s.logger.Printf("server listening in port: %v", server.Addr)

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatal("starting server", err)
		}
	}()

	<-stop
	s.shutdown(server)
}

func (s *Server) shutdown(server *http.Server) {
	s.logger.Printf("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.logger.Printf("ERR: shutting down server: %v", err)
	} else {
		s.logger.Println("Server shutted down succefully")
	}
}
