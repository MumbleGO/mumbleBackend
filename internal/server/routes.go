package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/inodinwetrust10/mumbleBackend/internal/database"
	"github.com/inodinwetrust10/mumbleBackend/utils"
)

func (s *Server) Handlers() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/user", utils.MakeHTTPHandleFunc(s.handleRegisterUser)).Methods("POST")
	return router
}

func (s *Server) handleRegisterUser(w http.ResponseWriter, r *http.Request) error {
	user := new(database.User)
	err := s.user.RegisterUser(user)
	if err != nil {
		return err
	}
	return nil
}
