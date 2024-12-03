package sockets

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func deleteRoom(roomid uuid.UUID) {
	Rooms.Delete(roomid)
	fmt.Println("room deleted")
}

func deleteUserFromRoom(roomid uuid.UUID, userid string) bool {
	// roomLock.Lock()
	// room := Rooms[roomid]
	// roomLock.Unlock()
	roomInterface, exists := Rooms.Load(roomid)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {

			room.mu.Lock()
			defer room.mu.Unlock()
			var lastIndex int = len(*room.nearByUserData) - 1
			for index, data := range *room.nearByUserData {
				if data.userId == userid {
					if index == lastIndex {
						*room.nearByUserData = (*room.nearByUserData)[:index]
					} else {
						*room.nearByUserData = append((*room.nearByUserData)[:index], (*room.nearByUserData)[index+1:]...)
					}
					fmt.Println("User left the room : ", userid)
				}
			}
			if len(*room.nearByUserData) <= 0 {
				deleteRoom(roomid)
			}
			return true
			// broad cast stayed users that this user has left the room!
			// is been done in handler file

			// type send_user_data struct {
			// 	Type   string `json:"type"`
			// 	UserId string `json:"userid"`
			// }
			// var send_user_data_json send_user_data = send_user_data{
			// 	Type:   "user_left_message",
			// 	UserId: userid,
			// }
			// jsonData, err := json.Marshal(send_user_data_json)
			// if err != nil {
			// 	log.Println("Could not marshal json")
			// 	return false
			// }
			// conn := e.connection[fd]
			// sendMessage(jsonData, &conn)
		}
	} else {
		fmt.Println("Room doesnt exists")
	}
	return false
}
func Informing_deletion(roomId uuid.UUID, userId string) {
	roomInterface, ok := Rooms.Load(roomId)
	if ok {
		room, ok := roomInterface.(*Room)
		if ok {
			var wg sync.WaitGroup
			for _, data := range *room.nearByUserData {
				wg.Add(1)
				go func() {
					defer wg.Done()
					go Epoller.InformDeletionResponse(data.fd, userId)
				}()
			}
			wg.Wait()
		}
	}
}
func (e *Epoll) join_user_room_again(roomId uuid.UUID, fd int) {
	fmt.Println("trying to let user join again!")
	roomInterface, ok := Rooms.Load(roomId)
	conn := e.Connection[fd]
	if conn == nil {
		return
	}
	if ok {
		room, ok := roomInterface.(*Room)
		if ok {
			// room.mu.Lock()
			// defer room.mu.Unlock()
			// var addNearUser nearbyUsers = nearbyUsers{
			// 	userId:    userId,
			// 	longitude: longitude,
			// 	latitude:  latitude,
			// 	fd:        fd,
			// }
			// *room.nearByUserData = append(*room.nearByUserData, addNearUser)
			fmt.Println("getting json to join room again")
			jsonData := MkNearUsersJsonData(room.nearByUserData, roomId)
			conn := e.Connection[fd]

			if jsonData != nil && conn != nil {
				fmt.Println("sending data to join room again")
				sendMessage(jsonData, &conn)
			}
		} else {
			join_error_respone(conn)
		}
	} else {
		join_error_respone(conn)
	}
}
func (e *Epoll) join_room_again(roomId uuid.UUID, userId string, longitude float64, latitude float64, fd int) {
	// roomLock.Lock()
	// room := Rooms[roomId]
	// roomLock.Unlock()
	fmt.Println("trying to let user join again!")
	roomInterface, exists := Rooms.Load(roomId)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			// check if user already exists
			var user_already_inroom bool = false
			if *room.nearByUserData != nil {
				for _, data := range *room.nearByUserData {
					if data.userId == userId {
						user_already_inroom = true
					}
				}
			}

			if !user_already_inroom {
				if len(*room.nearByUserData) != 0 {
					// adding user back into the nearUser array
					var addNearUser nearbyUsers = nearbyUsers{
						userId:    userId,
						longitude: longitude,
						latitude:  latitude,
						fd:        fd,
					}
					*room.nearByUserData = append(*room.nearByUserData, addNearUser)
					fmt.Println("getting json to join room again")
					jsonData := MkNearUsersJsonData(room.nearByUserData, roomId)
					conn := e.Connection[fd]

					if jsonData != nil && conn != nil {
						fmt.Println("sending data to join room again")
						sendMessage(jsonData, &conn)
					}
				} else {
					deleteRoom(roomId)
				}
				// all users left
			}
		}
	}
}
