package sockets

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	radii  float64 = 100
	client *redis.Client
	// ctx    context.Context
)

func init() {
	// ctx = context.Background()
	opt, _ := redis.ParseURL("rediss://default:AWuGAAIjcDE4ZmU4MmI2NDkzMjA0NDhlYWMxMzlmYTZkMGNhZmZmMnAxMA@improved-giraffe-27526.upstash.io:6379")
	client = redis.NewClient(opt)
}

func addUserGeoLocation(longitude float64, latitude float64, userId string) {
	_, err := client.GeoAdd(context.Background(), "user_locations", &redis.GeoLocation{
		Name:      userId,
		Longitude: longitude, // Longitude first
		Latitude:  latitude,  // Latitude second
	}).Result()
	if err != nil {
		log.Printf("Error adding user location: %v\n", err)
	}
}

func addDestinationGeoLocation(longitude float64, latitude float64, userId string) {
	_, err := client.GeoAdd(context.Background(), "user_destination", &redis.GeoLocation{
		Name:      userId,
		Longitude: longitude,
		Latitude:  latitude,
	}).Result()
	if err != nil {
		log.Println("error adding user destination : ", err)
	}
}

// func getUserGeoLocation(userId string) ([]*redis.GeoPos, error) {
// 	userKey := fmt.Sprintf("user_location:%s", userId)
// 	val, err := client.GeoPos(ctx, userKey, userId).Result()
// 	if err != nil {
// 		fmt.Println("Could not get user location :", userId)
// 		return nil, err
// 	}
// 	return val, nil
// }

func getNearbyUsers(longitude float64, latitude float64, destination_longitude float64, destination_latitude float64) ([]redis.GeoLocation, error) {
	current_location_val, err := client.GeoRadius(context.Background(), "user_locations", longitude, latitude, &redis.GeoRadiusQuery{
		Radius:    radii,
		Unit:      "km",
		WithCoord: true,
		Count:     10,
	}).Result()

	if err != nil {
		log.Printf("Error fetching nearby users: %v\n", err)
		return nil, err
	}
	destination_val, err := client.GeoRadius(context.Background(), "user_destination", destination_longitude, destination_latitude, &redis.GeoRadiusQuery{
		Radius:    radii,
		Unit:      "km",
		WithCoord: true,
		Count:     10,
	}).Result()
	if err != nil {
		log.Printf("Error fetching destination users: %v\n", err)
		return nil, err
	}
	// get common location
	var common_users []redis.GeoLocation
	destinationMap := make(map[string]redis.GeoLocation)
	for _, loc := range destination_val {
		destinationMap[loc.Name] = loc
	}

	for _, loc := range current_location_val {
		if _, ok := destinationMap[loc.Name]; ok {
			common_users = append(common_users, loc)
		}
	}

	// Print found users for debugging
	for _, location := range common_users {
		fmt.Printf("Nearby user: %v, Location: %v, %v\n", location.Name, location.Longitude, location.Latitude)
	}
	return common_users, nil
}
