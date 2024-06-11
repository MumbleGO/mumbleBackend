package server

import (
	"fmt"
	"net/http"

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

func (s *Server) Run() error {
	router := s.Handlers()
	server := &http.Server{
		Addr:    s.listenAddr,
		Handler: router,
	}
	fmt.Print("listening to port ")
	return server.ListenAndServe()
}
