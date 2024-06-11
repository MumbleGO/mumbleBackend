package database

import (
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MessagePlain struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
	Timestamp  string
	IsFile     bool
	FilePath   string
}

type PostgresMessage struct {
	db *gorm.DB
}
type MessageOperations interface {
	CreateMessage(*MessagePlain) error
}

func NewPostgresMessage() (*PostgresMessage, error) {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_STRING")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	connection := &PostgresMessage{db: conn}
	return connection, err
}

func (mess *PostgresMessage) CreateMessage(m *MessagePlain) error {
	return nil
}
