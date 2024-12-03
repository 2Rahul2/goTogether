package sockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func PollsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	err = Epoller.Add(conn)
	if err != nil {
		log.Printf("failed to add connection %v", err)
		conn.Close()
	}
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	go func() {
		defer conn.Close()
		log.Println("client connected")

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				log.Printf("Read error :%v", err)
				break
			}

			log.Printf("Message recv : %v", string(msg))
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}

	}()
}

// var all_user []string = {""}
func Start() {
	for {
		connections, err := Epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
		}

		for _, conn := range connections {
			if conn == nil {
				break
			}
			if msg, _, err := wsutil.ReadClientData(conn); err != nil {
				if err := Epoller.Remove(conn); err != nil {
					log.Printf("failed to remove epoll %v", err)
				}
				fmt.Println("removing user connection")
				conn.Close()
			} else {
				go func() {
					var message Message
					err := json.Unmarshal(msg, &message)
					if err != nil {
						fmt.Println("Error marshalling")
					}
					switch message.Type {
					case "user_left":
						fmt.Println("user left with room id :")
					case "send_message":
						// var Data struct {
						// 	RoomId   uuid.UUID `json:"roomid"`
						// 	Message  string    `json:"message"`
						// 	Username string    `json:"username"`
						// }
						var Data ChatMessageData
						err := json.Unmarshal(message.Data, &Data)
						fmt.Println("got chat message")
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							return
						}
						fmt.Println("no error marshaling chat message")
						Data.sendMessages()
					case "cancel_search":
						fmt.Println("user has sent cancel request!")
						var Data struct {
							UserId string `json:"userId"`
						}
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							return
						}
						cancelSearchRequest(Data.UserId)
						// response is sent for this use case
					case "room_deletion":
						var Data struct {
							RoomId uuid.UUID `json:"roomid"`
						}
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							return
						}
						deleteRoom(Data.RoomId)
					case "allusers":
						displayAllusers()
					case "add_user_existroom":
						fmt.Println("got request to check rooms for user")
						var Data struct {
							UserId string    `json:"userid"`
							RoomId uuid.UUID `json:"roomid"`
							Lat    float64   `json:"lat"`
							Lng    float64   `json:"lng"`
						}
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							return
						}
						var fd int = websocketFd(conn)
						Epoller.join_room_again(Data.RoomId, Data.UserId, Data.Lng, Data.Lat, fd)
					case "commonpoint_location":
						var Data Commonpoint_location
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							return
						}
						Data.SendCommonLocation()
					case "commonpoint_select_location":
						var Data Commonpoint_select_location
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							return
						}
						Data.SendSelectCommonLocation()
					case "delete_select_location":
						var Data Commonpoint_deletion
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							return
						}
						Data.DeleteCommonLocation()
					case "join_again":
						var Data struct {
							RoomId uuid.UUID `json:"roomid"`
						}
						var fd int = websocketFd(conn)
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							join_error_respone(conn)
							return
						}
						Epoller.join_user_room_again(Data.RoomId, fd)
					case "delete_user_room":
						var Data struct {
							UserId string    `json:"userid"`
							RoomId uuid.UUID `json:"roomid"`
						}
						var fd int = websocketFd(conn)
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							return
						}
						var deleteStatus bool = deleteUserFromRoom(Data.RoomId, Data.UserId)
						Epoller.SendDeleteUserFromRoomResponse(fd, deleteStatus)
						if deleteStatus {
							Informing_deletion(Data.RoomId, Data.UserId)
						}
					case "startSearching":
						var Data struct {
							Username        string  `json:"username"`
							Userid          string  `json:"userid"`
							User_lat        float64 `json:"user_lat"`
							User_lng        float64 `json:"user_lng"`
							Destination_lat float64 `json:"destination_lat"`
							Destination_lng float64 `json:"destination_lng"`
						}

						var fd int = websocketFd(conn)
						fmt.Println("incomming data : ", message.Data)
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
							sendMessage([]byte{}, &conn)
							return
						}
						fmt.Println("FDDD :  ", fd, Data.Userid)
						addUsers(string(Data.Userid), Data.Username, Data.User_lng, Data.User_lat, fd, Data.Destination_lng, Data.Destination_lat)
						userFriends, statusId, roomId := searchUsers(string(Data.Userid), Data.Username, Data.User_lng, Data.User_lat, fd, Data.Destination_lng, Data.Destination_lat)
						if statusId == 0 {
							Epoller.SendNoNearybyUserFoundResponse(fd)
							log.Println("Could not find users...")
						} else if statusId == 2 {
							log.Println("will get connected by host in a moment")
						} else {
							log.Println("this is host , sending connection status to other clients")
							Epoller.broadCastRoom(userFriends, roomId)
						}
						fmt.Println(userFriends)
						// response is sent for this use case
					case "getRooms":
						getRoomDetails()
					case "deleteUser":
						var Data struct {
							UserId string `json:"userid"`
						}
						err := json.Unmarshal(message.Data, &Data)
						if err != nil {
							fmt.Println("Error marshalling data 'Data' ")
						}
						removeUsers(string(Data.UserId))
					default:
						log.Printf("Unknown message type %s", message.Type)
					}

				}()
				// log.Printf("msg : %s", string(msg))

				// go startJoining(string(msg) , )
				// err = wsutil.WriteServerMessage(conn, op, msg)
				// if err != nil {
				// 	log.Printf("Write error: %v", err)
				// 	break
				// }

			}
		}
	}
}
