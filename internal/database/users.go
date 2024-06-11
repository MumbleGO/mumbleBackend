package database

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/inodinwetrust10/mumbleBackend/utils"
)

type UserPlain struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	FullName        string    `json:"fullname"`
	Password        string    `json:"password"`
	ConfirmPassword string    `json:"confirmPassword"`
	Gender          string    `json:"gender"`
	ProfilePic      string    `json:"profile_picture"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PostgresUser struct {
	db *gorm.DB
}

type UserOperations interface {
	SignUp(*UserPlain, http.ResponseWriter) error
}

func NewPostgresUser() (*PostgresUser, error) {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_STRING")), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := conn.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Fatal("Failed to enable UUID extension:", err)
	}

	// Create gender enum type if it doesn't exist
	if err := conn.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender') THEN
			CREATE TYPE gender AS ENUM ('male', 'female');
		END IF;
	END $$;`).Error; err != nil {
		log.Fatal("Failed to create gender enum type:", err)
	}
	// Perform auto-migration
	err = conn.AutoMigrate(&User{}, &Conversation{}, &Message{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
	}

	log.Println("Database migration completed successfully.")
	connection := &PostgresUser{db: conn}
	return connection, err
}

// ////////////////////////////////////////////////////////////////////////////////////
func Validate(u *UserPlain) error {
	if u.ID == "" {
		return errors.New("id cannot be empty")
	}
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

// /////////////////////////////////////////////////////////////////////////////////////
func (u *PostgresUser) SignUp(user *UserPlain, w http.ResponseWriter) error {
	var existingUser User
	if err := u.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return utils.WriteJson(
			w,
			http.StatusBadRequest,
			utils.ApiError{ErrorMessage: "user already exists"},
		)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	fmt.Println(hashedPassword)
	if err != nil {
		return err
	}
	return nil
}
