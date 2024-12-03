package sockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

// go tool pprof http://localhost:6060/debug/pprof/heap?  svg > profile_output.svg
type Epoll struct {
	Fd         int
	Connection map[int]net.Conn
	Lock       *sync.RWMutex
	OpCode     map[int]ws.OpCode
}

var Epoller *Epoll

func MkEpoll() (*Epoll, error) {
	Fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		Fd:         Fd,
		Lock:       &sync.RWMutex{},
		Connection: make(map[int]net.Conn),
		OpCode:     make(map[int]ws.OpCode),
	}, nil
}

func (e *Epoll) Add(conn net.Conn) error {
	Fd := websocketFd(conn)
	err := unix.EpollCtl(e.Fd, syscall.EPOLL_CTL_ADD, Fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(Fd)})
	if err != nil {
		return err
	}
	e.Lock.Lock()
	defer e.Lock.Unlock()
	e.Connection[Fd] = conn
	return nil
}

func (e *Epoll) Remove(conn net.Conn) error {
	Fd := int(websocketFd(conn))
	err := unix.EpollCtl(e.Fd, syscall.EPOLL_CTL_DEL, Fd, nil)
	if err != nil {
		return err
	}
	e.Lock.Lock()
	defer e.Lock.Unlock()
	delete(e.Connection, Fd)
	return nil
}

func (e *Epoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.Fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.Lock.RLock()
	defer e.Lock.RUnlock()
	var Connection []net.Conn
	for i := 0; i < n; i++ {
		conn := e.Connection[int(events[i].Fd)]
		Connection = append(Connection, conn)
	}
	return Connection, nil
}

func (e *Epoll) broadCastRoom(notify_user *[]nearbyUsers, roomid uuid.UUID) {
	fmt.Println("GONNNA BROAD CASTT NOWWWW!!!!")
	fmt.Println("USERS AREE : ", notify_user)
	jsonData := MkNearUsersJsonData(notify_user, roomid)
	log.Println(jsonData)
	log.Println("NOTIFY USER VALUES :   ", *notify_user)
	for _, data := range *notify_user {
		log.Println("SENDING DATA TO THE USER WITH OP CODE :")
		// log.Println("OP CODE :", *data.opCode, data.opCode, "FD :  ", data.Fd)
		conn := e.Connection[data.fd]
		if conn != nil {
			sendMessage(jsonData, &conn)
		}
	}
}
func MkNearUsersJsonData(notify_user *[]nearbyUsers, roomid uuid.UUID) []byte {
	type NearbyUsers struct {
		Username  string  `json:"username"`
		UserId    string  `json:"userid"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	type SendData struct {
		Type   string        `json:"type"`
		Result string        `json:"result"`
		Users  []NearbyUsers `json:"users"`
		RoomId uuid.UUID     `json:"roomid"`
	}
	var NearUsers []NearbyUsers
	for _, data := range *notify_user {
		NearUsers = append(NearUsers, NearbyUsers{
			Username:  data.username,
			UserId:    data.userId,
			Longitude: data.longitude,
			Latitude:  data.latitude,
		})
	}
	send_user_data := SendData{
		Type:   "found_users",
		Result: "success",
		Users:  NearUsers,
		RoomId: roomid,
	}

	jsonData, err := json.Marshal(send_user_data)
	if err != nil {
		log.Println("COULD NOT MARSHALL DATAA NOOOOO!!!!!!")
		return nil
		// reutrn 404 error :could not parse json data
	}
	return jsonData
}
func sendMessage(jsonData []byte, conn *net.Conn) {
	if conn != nil {
		err := wsutil.WriteServerMessage(*conn, ws.OpText, jsonData)
		if err != nil {
			fmt.Println("error sending json data to user")
		}
	}
}

func websocketFd(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
