package sockets

import (
	"encoding/json"

	"github.com/google/uuid"
)

type UserLocation struct {
	Roomid uuid.UUID `json:"roomid"`
	Userid string    `json:"userid"`
	Lat    float64   `json:"lat"`
	Lng    float64   `json:"lng"`
}

func (UL *UserLocation) JsonUpdateLocation() []byte {
	type Data struct {
		Type   string  `json:"type"`
		Userid string  `json:"userid"`
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
	}

	var data Data = Data{
		Type:   "sendmy_location",
		Userid: UL.Userid,
		Lat:    UL.Lat,
		Lng:    UL.Lng,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return jsonData
}

func (UL *UserLocation) UpdateLocation() {
	roomInterface, exists := Rooms.Load(UL.Roomid)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			jsonData := UL.JsonUpdateLocation()
			for _, data := range *room.nearByUserData {
				Epoller.broadCastMessage(data.fd, jsonData)
			}
		}
	}
}
