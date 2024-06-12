package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inodinwetrust10/mumbleBackend/internal/database"
)

type Server struct {
	listenAddr string
	user       database.UserOperations
	messages   database.MessageOperations
}

func NewServer(listenAddr string, u database.UserOperations, m database.MessageOperations) *Server {
	return &Server{
		listenAddr: listenAddr,
		user:       u,
		messages:   m,
	}
}

func (s *Server) Run() {
	router := s.Handlers()
	server := &http.Server{
		Addr:    s.listenAddr,
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting the server on port 3000")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server Error: %v", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("Shutting down the server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}
