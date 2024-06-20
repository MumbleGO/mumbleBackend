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
	router.Use(middleware.CORSMiddleware)

	router.HandleFunc("/api/auth/signup", utils.MakeHTTPHandleFunc(s.handleSignUp)).Methods("POST")
	router.HandleFunc("/api/auth/login", utils.MakeHTTPHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/api/auth/logout", utils.MakeHTTPHandleFunc(s.handleLogout)).Methods("POST")
	router.Handle("/api/auth/me", middleware.AuthMiddleware(utils.MakeHTTPHandleFunc(s.handleMe))).
		Methods("GET")

	router.Handle("/api/message/conversations", middleware.AuthMiddleware(utils.MakeHTTPHandleFunc(s.handleGetUserForSidebar))).
		Methods("GET")

	router.Handle("/api/message/send/{id}", middleware.AuthMiddleware(utils.MakeHTTPHandleFunc(s.handleSendMessage))).
		Methods("POST")

	router.Handle("/api/message/{id}", middleware.AuthMiddleware(utils.MakeHTTPHandleFunc(s.handleGetMessage))).
		Methods("GET")
	router.HandleFunc("/ws", s.handleWS)
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
	id := r.Context().Value("id").(string)
	err := s.user.GetMe(id, w)
	return err
}

// ////////////////////////////////////////////////////////////////////////////////////
func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) error {
	receiverID, senderID := getID(r)
	message, err := database.DecodeMessage(r)
	if err != nil {
		return err
	}
	err = s.messages.SendMessage(message, senderID, receiverID, w)
	if err != nil {
		return err
	}

	return nil
}

/////////////////////////////////////////////////////////////////////////////////////

func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request) error {
	userToChatID, senderID := getID(r)
	err := s.messages.GetMessage(userToChatID, senderID, w)
	if err != nil {
		return err
	}
	return nil
}

// ////////////////////////////////////////////////////////////////////////////////////
func (s *Server) handleGetUserForSidebar(w http.ResponseWriter, r *http.Request) error {
	_, authUser := getID(r)
	err := s.messages.GetUserForSidebar(authUser, w)
	if err != nil {
		return err
	}
	return nil
}

func getID(r *http.Request) (string, string) {
	userToChatID := mux.Vars(r)["id"]
	senderID := r.Context().Value("id").(string)
	return userToChatID, senderID
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	database.HandleWebSocket(w, r)
}
