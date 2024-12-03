package sockets

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ChatMessageData struct {
	RoomId       uuid.UUID `json:"roomId"`
	Message      string    `json:"message"`
	Username     string    `json:"username"`
	Senderuserid string    `json:"senderuserid"`
}

func (chatData *ChatMessageData) jsonChatMessage() []byte {
	type Message struct {
		Type         string `json:"type"`
		Msg          string `json:"msg"`
		Username     string `json:"username"`
		Senderuserid string `json:"senderuserid"`
	}
	var messageData Message = Message{
		Type:         "message_recv",
		Msg:          chatData.Message,
		Username:     chatData.Username,
		Senderuserid: chatData.Senderuserid,
	}
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		// send bad response too all users
		return nil
	}
	return jsonData

}

func (chatData *ChatMessageData) sendMessages() {
	roomInterface, exists := Rooms.Load(chatData.RoomId)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			jsonData := chatData.jsonChatMessage()
			for _, data := range *room.nearByUserData {
				Epoller.broadCastMessage(data.fd, jsonData)
			}
		}
	}
}
