package database

import (
	"encoding/json"
	"log"
	"net/http"

	"gorm.io/gorm"

	"github.com/inodinwetrust10/mumbleBackend/utils"
)

type MessageOperations interface {
	SendMessage(*MessagePlain, string, string, http.ResponseWriter) error
	GetMessage(string, string, http.ResponseWriter) error
	GetUserForSidebar(string, http.ResponseWriter) error
}

func NewPostgresMessage() (*PostgresMessage, error) {
	conn, err := ExpoDB()
	if err != nil {
		log.Fatal(err)
	}
	connection := &PostgresMessage{db: conn}
	return connection, err
}

func (m *PostgresMessage) SendMessage(
	mess *MessagePlain,
	senderId string,
	receiverId string,
	w http.ResponseWriter,
) error {
	var conversation Conversation

	subQuery := m.db.Table("conversation_participants").
		Select("conversation_id").
		Where("user_id IN (?, ?)", senderId, receiverId).
		Group("conversation_id").
		Having("COUNT(DISTINCT user_id) = ?", 2)

	err := m.db.Where("id IN (?)", subQuery).
		First(&conversation).Error

	if err == gorm.ErrRecordNotFound {
		conversation = Conversation{}
		err = m.db.Create(&conversation).Error
		if err != nil {
			return err
		}

		sender := User{ID: senderId}
		receiver := User{ID: receiverId}

		err = m.db.Model(&conversation).Association("Participants").Append(&sender, &receiver)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	newMessage := Message{
		SenderID:       senderId,
		Body:           mess.Content,
		ConversationID: conversation.ID,
	}
	err = m.db.Create(&newMessage).Error
	if err != nil {
		return err
	}

	err = m.db.Preload("Participants").
		Preload("Messages").
		First(&conversation, "id = ?", conversation.ID).
		Error
	if err != nil {
		return err
	}

	return utils.WriteJson(w, http.StatusCreated, &newMessage)
}

// /////////////////////////////////////////////////////////////////////////////////////
func (m *PostgresMessage) GetMessage(toChat string, senderID string, w http.ResponseWriter) error {
	var conversation Conversation

	err := m.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id IN (?)", []string{senderID, toChat}).
		Group("conversations.id").
		Find(&conversation).Error
	if err != nil {
		return utils.WriteJson(
			w,
			http.StatusInternalServerError,
			utils.ApiError{ErrorMessage: "Failed to fetch the conversation"},
		)
	}

	if conversation.ID == "" {
		return utils.WriteJson(w, http.StatusOK, []interface{}{})
	}

	return utils.WriteJson(w, http.StatusOK, conversation.Messages)
}

// ////////////////////////////////////////////////////////////////////////////////////
func (m *PostgresMessage) GetUserForSidebar(authUser string, w http.ResponseWriter) error {
	var users []UserInfo

	err := m.db.Model(&User{}).
		Select("id, full_name, profile_pic").
		Where("id != ?", authUser).
		Find(&users).Error
	if err != nil {
		return utils.WriteJson(
			w,
			http.StatusInternalServerError,
			utils.ApiError{ErrorMessage: "Failed to fetch users"},
		)
	}

	return utils.WriteJson(w, http.StatusOK, users)
}

// ////////////////////////////////////////////////////////////////////////////////////
func DecodeMessage(r *http.Request) (*MessagePlain, error) {
	message := new(MessagePlain)
	err := json.NewDecoder(r.Body).Decode(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
