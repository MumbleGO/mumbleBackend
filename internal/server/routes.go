package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/inodinwetrust10/mumbleBackend/internal/database"
	"github.com/inodinwetrust10/mumbleBackend/internal/middleware"
	"github.com/inodinwetrust10/mumbleBackend/utils"
)

func (s *Server) Handlers() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/auth/signup", utils.MakeHTTPHandleFunc(s.handleSignUp)).Methods("POST")
	router.HandleFunc("/api/auth/login", utils.MakeHTTPHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/api/auth/logout", utils.MakeHTTPHandleFunc(s.handleLogout)).Methods("POST")
	router.Handle("/api/auth/me", middleware.AuthMiddleware(utils.MakeHTTPHandleFunc(s.handleMe))).
		Methods("GET")
	return router
}

// /////////////////////////////////////////////////////////////////////////////////////
func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) error {
	user, err := database.DecodeUser(r)
	if err != nil {
		return err
	}
	err = database.ValidateUser(user)
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

// /////////////////////////////////////////////////////////////////////////////////////
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) error {
	user, err := database.DecodeUser(r)
	if err != nil {
		return err
	}
	err = s.user.Login(user, w)
	if err != nil {
		return err
	}

	return nil
}

// /////////////////////////////////////////////////////////////////////////////////////
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) error {
	err := s.user.Logout(w)
	if err != nil {
		return err
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) error {
	username := r.Context().Value("username").(string)
	err := s.user.GetMe(username, w)
	return err
}
