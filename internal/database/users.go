package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/inodinwetrust10/mumbleBackend/utils"
)

type UserOperations interface {
	SignUp(*UserPlain, http.ResponseWriter) error
	Login(*UserPlain, http.ResponseWriter) error
	Logout(http.ResponseWriter) error
}

func NewPostgresUser() (*PostgresUser, error) {
	conn, err := ExpoDB()
	if err != nil {
		return nil, err
	}
	connection := &PostgresUser{db: conn}
	return connection, err
}

// ////////////////////////////////////////////////////////////////////////////////////

func ValidateUser(u *UserPlain) error {
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}
	if u.FullName == "" {
		return errors.New("full_name cannot be empty")
	}
	if u.Password == "" {
		return errors.New("password cannot be empty")
	}
	if u.ConfirmPassword == "" {
		return errors.New("password cannot be empty")
	}
	if u.Gender == "" {
		return errors.New("gender cannot be empty")
	}
	if u.ConfirmPassword != u.Password {
		return errors.New("passwords are not same")
	}
	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////
func (u *PostgresUser) SignUp(user *UserPlain, w http.ResponseWriter) error {
	var existingUser User
	err := u.db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		// User already exists
		return utils.WriteJson(
			w,
			http.StatusBadRequest,
			utils.ApiError{ErrorMessage: "username already taken"},
		)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// An error other than "record not found" occurred
		return err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	// Set profile picture URL based on gender
	var userProfilePic string
	if user.Gender == "male" {
		userProfilePic = fmt.Sprintf(
			"https://avatar.iran.liara.run/public/boy?username=%s",
			user.Username,
		)
	} else {
		userProfilePic = fmt.Sprintf(
			"https://avatar.iran.liara.run/public/girl?username=%s",
			user.Username,
		)
	}

	// Create a new user record
	newUser := &User{
		FullName:   user.FullName,
		Username:   user.Username,
		Password:   string(hashedPassword),
		Gender:     Gender(user.Gender),
		ProfilePic: userProfilePic,
	}

	// Save the new user to the database
	if err := u.db.Create(newUser).Error; err != nil {
		return err
	}
	err = utils.GenerateJWT(newUser.Username, w)
	if err != nil {
		return err
	}

	return utils.WriteJson(
		w,
		http.StatusCreated,
		&UserPlain{
			ID:         newUser.ID,
			FullName:   newUser.FullName,
			Username:   newUser.Username,
			ProfilePic: newUser.ProfilePic,
			Gender:     string(newUser.Gender),
		},
	)
}

// ////////////////////////////////////////////////////////////////////////////////////
func (pu *PostgresUser) Login(u *UserPlain, w http.ResponseWriter) error {
	var existingUser User
	err := pu.db.Where("username = ?", u.Username).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// User does not exists
		return utils.WriteJson(
			w,
			http.StatusNotFound,
			utils.ApiError{ErrorMessage: "No account found with this username"},
		)
	} else if err != nil {
		return utils.WriteJson(w, http.StatusNotFound, utils.ApiError{ErrorMessage: "Error in fetching the user"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(u.Password))
	if err != nil {
		return utils.WriteJson(
			w,
			http.StatusUnauthorized,
			utils.ApiError{ErrorMessage: "Wrong username or password"},
		)
	}

	err = utils.GenerateJWT(u.Username, w)
	if err != nil {
		return err
	}
	return utils.WriteJson(
		w,
		http.StatusOK,
		&UserPlain{
			ID:         existingUser.ID,
			FullName:   existingUser.FullName,
			Username:   existingUser.Username,
			ProfilePic: existingUser.ProfilePic,
		},
	)
}

// ///////////////////////////////////////////////////////////////////////////////////
func (u *PostgresUser) Logout(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
	return utils.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "logged out successfully"},
	)
}

// ////////////////////////////////////////////////////////////////////////////////////
func DecodeUser(r *http.Request) (*UserPlain, error) {
	user := new(UserPlain)
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
