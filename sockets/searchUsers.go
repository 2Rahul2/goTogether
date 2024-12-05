package sockets

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type nearbyUsers struct {
	username  string
	userId    string
	longitude float64
	latitude  float64
	fd        int
}

func cancelSearchRequest(userId string) {
	userLock.Lock()
	user := users[userId]
	defer userLock.Unlock()
	delete(users, userId)
	if user != nil {
		Epoller.SendCancelResponse(user.fd)
	}
}
func checkForCancel(userId string) bool {
	userLock.Lock()
	defer userLock.Unlock()
	user := users[userId]
	return user == nil
}
func searchUserInterval(stopChan chan bool, userId string, longitude float64, latitude float64, destination_longitude float64, destination_latitude float64) *[]nearbyUsers {
	fmt.Println("started searching...")
	var totalSearchCount int = 0
	// Create a ticker to call the function at regular intervals
	ticker := time.NewTicker(3 * time.Second)
	// defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping search...")
			return nil
		case <-ticker.C:
			if checkForCancel(userId) {
				stopChan <- true
				return nil
			}
			var userInfoArray []nearbyUsers
			val, err := getNearbyUsers(longitude, latitude, destination_longitude, destination_latitude)
			if err != nil {
				log.Println(err)
				return nil
			}
			userLock.Lock()
			for _, locationData := range val {
				user := users[locationData.Name]
				if user != nil {
					userinfo := nearbyUsers{
						user.username,
						locationData.Name,
						locationData.Longitude,
						locationData.Latitude,
						user.fd,
					}
					userInfoArray = append(userInfoArray, userinfo)
				} else {
				}
			}
			userLock.Unlock()

			if len(userInfoArray) > 1 {
				stopChan <- true
				return &userInfoArray
			}
			totalSearchCount += 1
			if totalSearchCount > 6 {
				stopChan <- true
				return nil
			}
			fmt.Println("No user found, will search again...")
		}
	}
}

func searchUsers(userId string, username string, user_longitude float64, user_latitude float64, fd int, destination_longitude float64, destination_latitude float64) (*[]nearbyUsers, uint, uuid.UUID) {
	fmt.Println("searching users...")
	stopChan := make(chan bool)

	// Call searchUserInterval in a goroutine so it runs concurrently
	var nearby_users_array *[]nearbyUsers
	go func() {
		nearby_users_array = searchUserInterval(stopChan, userId, user_longitude, user_latitude, destination_longitude, destination_latitude)
	}()

	// Wait for searchUserInterval to finish or return results
	<-stopChan

	if nearby_users_array == nil || len(*nearby_users_array) < 1 {
		// No users found; return a bad response to the user...
		return nil, 0, uuid.Nil
	}

	nearby_Users, status, roomId := startJoiningRoom(userId, username, nearby_users_array, user_latitude, user_longitude, fd)

	return nearby_Users, status, roomId
}
