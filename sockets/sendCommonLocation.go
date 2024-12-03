package sockets

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Commonpoint_deletion struct {
	Userid      string    `json:"userid"`
	Roomid      uuid.UUID `json:"roomid"`
	Otheruserid string    `json:"otheruserid"`
}

type Commonpoint_select_location struct {
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Roomid      uuid.UUID `json:"roomid"`
	Userid      string    `json:"userid"`
	Address     string    `json:"address"`
	Otheruserid string    `json:"otheruserid"`
	Othername   string    `json:"othername"`
}

type Commonpoint_location struct {
	Lat     float64   `json:"lat"`
	Lng     float64   `json:"lng"`
	Roomid  uuid.UUID `json:"roomid"`
	Userid  string    `json:"userid"`
	Address string    `json:"address"`
}

func (CD *Commonpoint_deletion) jsonCommonDeletion() []byte {
	type Data struct {
		Type        string `json:"type"`
		Userid      string `json:"userid"`
		Otheruserid string `json:"otheruserid"`
	}

	jsonData, err := json.Marshal(Data{
		Type:        "delete_select_location",
		Userid:      CD.Userid,
		Otheruserid: CD.Otheruserid,
	})
	if err != nil {
		return nil
	}
	return jsonData
}

func (CD *Commonpoint_deletion) DeleteCommonLocation() {
	roomInterface, exists := Rooms.Load(CD.Roomid)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			jsonData := CD.jsonCommonDeletion()
			for _, data := range *room.nearByUserData {
				Epoller.broadCastMessage(data.fd, jsonData)
			}
		}
	}
}
func (CSL *Commonpoint_select_location) jsonCommonSelectLocation() []byte {
	type Data struct {
		Type        string  `json:"type"`
		Userid      string  `json:"userid"`
		Address     string  `json:"address"`
		Lng         float64 `json:"lng"`
		Lat         float64 `json:"lat"`
		Otheruserid string  `json:"otheruserid"`
		Othername   string  `json:"othername"`
	}
	jsonData, err := json.Marshal(Data{
		Type:        "commonpoint_location",
		Userid:      CSL.Userid,
		Address:     CSL.Address,
		Lng:         CSL.Lng,
		Lat:         CSL.Lat,
		Otheruserid: CSL.Otheruserid,
		Othername:   CSL.Othername,
	})
	if err != nil {
		return nil
	}
	return jsonData
}
func (CSL *Commonpoint_select_location) SendSelectCommonLocation() {
	roomInterface, exists := Rooms.Load(CSL.Roomid)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			jsonData := CSL.jsonCommonSelectLocation()
			for _, data := range *room.nearByUserData {
				Epoller.broadCastMessage(data.fd, jsonData)
			}
		}
	}
}

func (CL *Commonpoint_location) jsonCommonLocation() []byte {
	type Data struct {
		Type    string  `json:"type"`
		Userid  string  `json:"userid"`
		Address string  `json:"address"`
		Lng     float64 `json:"lng"`
		Lat     float64 `json:"lat"`
	}
	jsonData, err := json.Marshal(Data{
		Type:    "commonpoint_location",
		Userid:  CL.Userid,
		Address: CL.Address,
		Lng:     CL.Lng,
		Lat:     CL.Lat,
	})

	if err != nil {
		return nil
	}
	return jsonData
}

func (commonPoint *Commonpoint_location) SendCommonLocation() {
	roomInterface, exists := Rooms.Load(commonPoint.Roomid)
	if exists {
		room, ok := roomInterface.(*Room)
		if ok {
			room.mu.Lock()
			defer room.mu.Unlock()
			jsonData := commonPoint.jsonCommonLocation()
			for _, data := range *room.nearByUserData {
				Epoller.broadCastMessage(data.fd, jsonData)
			}
		}
	}
}
