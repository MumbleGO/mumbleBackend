package database

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/inodinwetrust10/mumbleBackend/utils"
)

type UserPlain struct {
	Username        string `json:"username"`
	FullName        string `json:"fullname"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Gender          string `json:"gender"`
}

type PostgresUser struct {
	db *gorm.DB
}

type UserOperations interface {
	SignUp(*UserPlain, http.ResponseWriter) error
}

func NewPostgresUser() (*PostgresUser, error) {
	conn, err := ExpoDB()
	if err != nil {
		return nil, err
	}
	connection := &PostgresUser{db: conn}
	return connection, err
}

// /////////////////////////////////////////////////////////////////////////////////////

func (u *PostgresUser) SignUp(user *UserPlain, w http.ResponseWriter) error {
	var existingUser User
	err := u.db.Where("username = ?", user.Username).First(&existingUser).Error
	if err == nil {
		// User already exists
		return utils.WriteJson(
			w,
			http.StatusBadRequest,
			utils.ApiError{ErrorMessage: "user already exists"},
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
	tokenString, err := utils.GenerateJWT(newUser.Username, os.Getenv("JWT_SECRET"))
	if err != nil {
		return err
	}

	// attaching the jwt token to the response w using cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
	return utils.WriteJson(w, http.StatusCreated, newUser)
}
