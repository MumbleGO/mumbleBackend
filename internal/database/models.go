package database

import (
	"time"
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
