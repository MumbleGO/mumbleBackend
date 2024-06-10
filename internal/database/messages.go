package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Message struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Content    string
	Timestamp  string
	IsFile     bool
	FilePath   string
}

type PostgresMessage struct {
	db *sql.DB
}
type Messages interface {
	CreateMessage(*Message) error
}

func NewPostgresMessage() (*PostgresMessage, error) {
	conn, err := sql.Open("postgres", os.Getenv("DB_STRING"))
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	connection := &PostgresMessage{db: conn}
	return connection, err
}

func (mess *PostgresMessage) CreateMessage(m *Message) error {
	return nil
}
