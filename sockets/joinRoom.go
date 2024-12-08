package sockets

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type User struct {
	username     string
	isjoinedRoom bool
	isLock       bool
	cond         *sync.Cond
	roomId       uuid.UUID
	fd           int
}

type Room struct {
	nearByUserData *[]nearbyUsers
	mu             sync.Mutex
}

// var Rooms = map[uuid.UUID]*Room{}
var Rooms sync.Map = sync.Map{}
var userLowerLimit int = 1
var userHighestLimit int = 2
var users = map[string]*User{}
var globalLock sync.Mutex
var userLock sync.RWMutex // For concurrent access to users map

var expirationDuration time.Duration = 2 * time.Hour

func getRoomDetails() {
	fmt.Println("displaying rooms details")
	Rooms.Range(func(key, value interface{}) bool {
		roomId, ok := key.(uuid.UUID)
		if !ok {
			return true
		}
		room, ok := value.(*Room)
		if !ok {
			return true
		}

		fmt.Printf("Room ID: %s\n", roomId)
		fmt.Println("User IDs:")
		for _, data := range *room.nearByUserData {
			fmt.Println(data.userId)
		}
		fmt.Println() // Empty line for better readability

		return true // Continue to the next item
	})
}

func displayAllusers() {
	fmt.Println(users)
}
func addUsers(userId string, username string, longitude float64, latitude float64, fd int, destination_longitude float64, destination_latitude float64) {
	userLock.Lock()
	defer userLock.Unlock()
	users[userId] = &User{
		username:     username,
		isjoinedRoom: false,
		isLock:       false,
		cond:         sync.NewCond(&sync.Mutex{}),
		fd:           fd,
	}
	addUserGeoLocation(longitude, latitude, userId)
	addDestinationGeoLocation(destination_longitude, destination_latitude, userId)
	log.Println("user added", users[userId])
}

// remove user if user isnt available for one minute PING PONG
// get room id from local storage...
// manual and automatic
func removeUsers(userId string) {
	userLock.Lock()
	defer userLock.Unlock()

	delete(users, userId)
	log.Println("deleted user :", userId)
}

// func releaseUsers(userId string) {
// 	userLock.Lock()
// 	defer userLock.Unlock()
// 	user := users[userId]
// 	if user != nil {
// 		user.cond.Broadcast()
// 		user.isLock = false
// 		user.isjoinedRoom = false
// 	}
// }

// 2 = user does not exist or in room already (exit the connection)
// 1 = joined room
// 0 = could not find other user(restart search)
func RoomExpiration(roomId uuid.UUID) {
	fmt.Println("gonna expire in ", expirationDuration)
	time.AfterFunc(expirationDuration, func() {
		// send room expired response to user
		val, ok := Rooms.Load(roomId)
		if ok {
			userLock.Lock()
			room := val.(*Room)
			for _, user := range *room.nearByUserData {
				fmt.Println("sending expire status to user : ", user.userId, user.fd)
				user_ := users[user.userId]
				if user_ != nil {
					delete(users, user.userId)
					//user.isjoinedRoom = false
				}
				Epoller.SendExpireRoomResponse(user.fd)
			}
			userLock.Unlock()

			Rooms.Delete(roomId)
		}
		Rooms.Delete(roomId)
	})
}

func startJoiningRoom(userId string, username string, alluserId *[]nearbyUsers, thisUserLat float64, thisUserLng float64, fd int) (*[]nearbyUsers, uint, uuid.UUID) {
	// Get the host user
	// globalLock.Lock()

	userLock.RLock()
	user := users[userId]
	userLock.RUnlock()

	if user == nil {
		log.Println("User not found:", userId)
		return nil, 0, uuid.Nil
	}

	fmt.Println("User Id : ", userId)
	user.cond.L.Lock()
	for user.isLock {
		user.cond.Wait()
	}
	user.cond.L.Unlock()
	globalLock.Lock()

	if user.isjoinedRoom {
		fmt.Println("User already in room")
		globalLock.Unlock()
		return nil, 2, uuid.Nil
	}
	user.cond.L.Lock()
	user.isLock = true

	// Start adding friends user of host and lock them

	// var wg sync.WaitGroup
	var readyTowait sync.WaitGroup
	toLockUsers := []string{}
	log.Println(userId, " group people are :", alluserId, len(*alluserId))
	var filtered_user *[]nearbyUsers = &[]nearbyUsers{}
	for _, data := range *alluserId {
		readyTowait.Add(1)
		go func(uid string) {
			userLock.RLock()
			otherUser := users[data.userId]
			userLock.RUnlock()
			if otherUser == nil {
				log.Println("User not found:", data.userId)
				// *alluserId = append((*alluserId)[:index], (*alluserId)[index+1:]...)
				readyTowait.Done()
				return
			}
			if !otherUser.isjoinedRoom && !otherUser.isLock {
				toLockUsers = append(toLockUsers, data.userId)
				otherUser.cond.L.Lock()
				otherUser.isLock = true
				log.Println(data.userId, " waiting to get unlocked")
				// otherUser.cond.Wait() // Waiting for signal to proceed
				log.Println(data.userId, " got unlocked")
				*filtered_user = append(*filtered_user, data)
				otherUser.cond.L.Unlock()
			}
			readyTowait.Done()
		}(data.userId)
	}
	fmt.Println("waiting for ready to wait")
	readyTowait.Wait()
	fmt.Println("waiting for ready to wait is completed")
	globalLock.Unlock()

	if len(*filtered_user) < userLowerLimit {
		// search for users again
		log.Println("not many users to create room for user :", userId, toLockUsers)
		return nil, 0, uuid.Nil
	}
	log.Println("LENGTH OF ALL USER ID  IS IS :      ", len(*filtered_user))
	if len(*filtered_user) > userHighestLimit {
		log.Println("reducing users!")
		for _, data := range (*filtered_user)[userHighestLimit:] {
			userLock.Lock()
			throwUser := users[data.userId]
			throwUser.cond.L.Lock()
			userLock.Unlock()
			throwUser.isLock = false
			throwUser.isjoinedRoom = false
			throwUser.cond.Broadcast()
			throwUser.cond.L.Unlock()
			// releaseUsers(data.userId)
		}
		*filtered_user = (*filtered_user)[:userHighestLimit]
		// *alluserId = append(*alluserId, (*alluserId)[:userHighestLimit]...)
	}
	log.Println("LENGTH OF ALL USER ID  IS IS :      ", len(*filtered_user), *filtered_user)

	// release condition wait and let them join in room
	var RoomId uuid.UUID = uuid.New()
	func() {
		userLock.Lock()
		for _, data := range *filtered_user {
			nearUser := users[data.userId]
			nearUser.isjoinedRoom = true
			nearUser.isLock = false
			nearUser.cond.Broadcast()
			nearUser.roomId = RoomId
		}
		userLock.Unlock()
	}()

	var thisUser = nearbyUsers{
		username:  username,
		userId:    userId,
		longitude: thisUserLng,
		latitude:  thisUserLat,
		fd:        fd,
	}
	*filtered_user = append(*filtered_user, thisUser)
	// roomLock.Lock()
	Rooms.Store(RoomId, &Room{
		nearByUserData: filtered_user,
		mu:             sync.Mutex{},
	})
	// add expiration timer
	log.Println("Added users in room")
	go RoomExpiration(RoomId)
	return filtered_user, 1, RoomId
}
