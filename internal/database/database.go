package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// /////////////////////////////////////////////////////////////////////////////////////
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

type UserPlain struct {
	ID              string `json:"id,omitempty"`
	Username        string `json:"username,omitempty"`
	FullName        string `json:"fullname,omitempty"`
	Password        string `json:"password,omitempty"`
	ConfirmPassword string `json:"confirmPassword,omitempty"`
	Gender          string `json:"gender,omitempty"`
	ProfilePic      string `json:"profile_picture,omitempty"`
}

type PostgresUser struct {
	db *gorm.DB
}

// /////////////////////////////////////////////////////////////////////////////////////
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

func Migrate() {
	db, err := ExpoDB()
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Fatal("Failed to enable UUID extension:", err)
	}
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
