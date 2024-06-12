package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/inodinwetrust10/mumbleBackend/internal/database"
	"github.com/inodinwetrust10/mumbleBackend/utils"
)

func (s *Server) Handlers() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/auth/signup", utils.MakeHTTPHandleFunc(s.handleSignUp)).Methods("POST")
	return router
}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) error {
	user := new(database.UserPlain)
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		return err
	}
	err = database.Validate(user)
	if err != nil {
		return utils.WriteJson(
			w,
			http.StatusBadRequest,
			utils.ApiError{ErrorMessage: "check the credentials again"},
		)
	}

	err = s.user.SignUp(user, w)
	if err != nil {
		return err
	}
	return nil
}
