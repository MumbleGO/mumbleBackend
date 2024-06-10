package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    string
	UpdatedAt    string
}

type PostgresUser struct {
	db *sql.DB
}

type Users interface {
	RegisterUser(*User) error
}

func NewPostgresUser() (*PostgresUser, error) {
	conn, err := sql.Open("postgres", os.Getenv("DB_STRING"))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	connection := &PostgresUser{db: conn}
	return connection, err
}

func (u *PostgresUser) RegisterUser(user *User) error {
	fmt.Println("hello from register")
	return nil
}
