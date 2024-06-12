package database

import (
	"errors"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID            string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username      string `gorm:"unique"`
	FullName      string
	Password      string
	Gender        Gender `gorm:"type:gender"`
	ProfilePic    string
	Conversations []Conversation `gorm:"many2many:user_conversations;constraint:OnDelete:CASCADE"`
	Messages      []Message      `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
}

// Conversation model
type Conversation struct {
	ID           string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Participants []User    `gorm:"many2many:conversation_participants;constraint:OnDelete:CASCADE"`
	Messages     []Message `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// Message model
type Message struct {
	ID             string       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ConversationID string       `gorm:"type:uuid;index;not null"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
	SenderID       string       `gorm:"type:uuid;index;not null"`
	Sender         User         `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Body           string
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// Gender type
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

func ExpoDB() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(os.Getenv("DB_STRING")), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, err
}

// ////////////////////////////////////////////////////////////////////////////////////
func Validate(u *UserPlain) error {
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

func Migrate() {
	db, err := ExpoDB()
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Fatal("Failed to enable UUID extension:", err)
	}

	// Create gender enum type if it doesn't exist
	if err := db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender') THEN
			CREATE TYPE gender AS ENUM ('male', 'female');
		END IF;
	END $$;`).Error; err != nil {
		log.Fatal("Failed to create gender enum type:", err)
	}
	// Perform auto-migration
	err = db.AutoMigrate(&User{}, &Conversation{}, &Message{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
	}

	log.Println("Database migration completed successfully.")
}