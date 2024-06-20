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
	notifyReceiver(receiverId, newMessage)
	messa := SendMessage{
		ID:             newMessage.ID,
		ConversationID: newMessage.ConversationID,
		SenderID:       newMessage.SenderID,
		Body:           newMessage.Body,
		CreatedAt:      newMessage.CreatedAt,
		UpdatedAt:      newMessage.UpdatedAt,
	}

	return utils.WriteJson(w, http.StatusCreated, messa)
}

// /////////////////////////////////////////////////////////////////////////////////////

func (m *PostgresMessage) GetMessage(toChat string, senderID string, w http.ResponseWriter) error {
	var conversation Conversation

	// Modified Query
	err := m.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).
		Joins("JOIN conversation_participants cp1 ON cp1.conversation_id = conversations.id").
		Joins("JOIN conversation_participants cp2 ON cp2.conversation_id = conversations.id").
		Where("cp1.user_id = ? AND cp2.user_id = ?", senderID, toChat).
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

	var messageArr []MessageType
	for _, mess := range conversation.Messages {
		message := MessageType{
			ID:        mess.ID,
			Body:      mess.Body,
			SenderID:  mess.SenderID,
			CreatedAt: mess.CreatedAt,
		}
		messageArr = append(messageArr, message)
	}
	return utils.WriteJson(w, http.StatusOK, messageArr)
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
