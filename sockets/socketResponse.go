package sockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func (e *Epoll) broadCastMessage(fd int, jsonData []byte) {
	conn := e.Connection[fd]
	if conn != nil {
		if jsonData != nil {
			err := wsutil.WriteServerMessage(conn, ws.OpText, jsonData)
			if err != nil {
				fmt.Println("lol didnt send message , some err :", err)
			}

		}
	}
}

func (e *Epoll) SendDeleteUserFromRoomResponse(fd int, success bool) {
	type send_user_data struct {
		Type   string `json:"type"`
		Result string `json:"result"`
	}
	var result string
	if success {
		result = "success"
	} else {
		result = "error"
	}
	var send_user_data_json send_user_data = send_user_data{
		Type:   "user_delete_status",
		Result: result,
	}
	jsonData, err := json.Marshal(send_user_data_json)
	if err != nil {
		log.Println("Could not marshal json")
		return
	}
	conn := e.Connection[fd]
	sendMessage(jsonData, &conn)
}
func (e *Epoll) InformDeletionResponse(fd int, userid string) {
	type send_user_data struct {
		Type   string `json:"type"`
		UserId string `json:"userid"`
	}
	var send_user_data_json send_user_data = send_user_data{
		Type:   "user_left_message",
		UserId: userid,
	}
	jsonData, err := json.Marshal(send_user_data_json)
	if err != nil {
		log.Println("Could not marshal json")
		return
	}
	conn := e.Connection[fd]
	sendMessage(jsonData, &conn)
}

func (e *Epoll) SendCancelResponse(fd int) {
	conn := e.Connection[fd]
	type cancelResponse struct {
		Type   string `json:"type"`
		Result string `json:"result"`
	}

	if conn == nil {
		return
	}
	cancel_response_data := cancelResponse{
		Type:   "search_cancel_request",
		Result: "success",
	}
	jsonData, err := json.Marshal(cancel_response_data)
	if err != nil {
		sendMessage([]byte{}, &conn)
		return
	}
	sendMessage(jsonData, &conn)
}

func (e *Epoll) SendExpireRoomResponse(fd int) {
	conn := e.Connection[fd]
	type expire_response struct {
		Type string `json:"type"`
	}
	if conn == nil {
		fmt.Println("conn is not available to send expire response  ~~~~~~~~~")
		return
	}
	jsonData, err := json.Marshal(expire_response{
		Type: "room_expired",
	})
	if err != nil {
		sendMessage([]byte{}, &conn)
		return
	}
	sendMessage(jsonData, &conn)

}

func (e *Epoll) SendNoNearybyUserFoundResponse(fd int) {
	conn := e.Connection[fd]
	if conn == nil {
		fmt.Println("no connection found fd: ", fd, e.Connection)
		return
	}
	type SendData struct {
		Type   string `json:"type"`
		Result string `json:"result"`
	}

	jsonData, err := json.Marshal(SendData{
		Type:   "found_users",
		Result: "not_found",
	})
	if err != nil {
		sendMessage([]byte{}, &conn)
		return
	}
	sendMessage(jsonData, &conn)
}

func join_error_respone(conn net.Conn) {
	type join_again_response struct {
		Type   string `json:"type"`
		Result string `json:"result"`
	}
	jsonData, err := json.Marshal(join_again_response{
		Type:   "found_users",
		Result: "error",
	})
	if err != nil {
		return
	}
	sendMessage(jsonData, &conn)
}
