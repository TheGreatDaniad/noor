package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type ConnectionPool []*net.Conn

func (cd ConnectionPool) RandomPick() *net.Conn {
	if len(cd) == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	// Generate a random index within the range of the slice length
	randomIndex := rand.Intn(len(cd))
	// Retrieve the random element from the slice
	return cd[randomIndex]
}

type Session struct {
	ID          uint16
	UserID      uint16
	RealIp      net.IP
	LocalIp     net.IP
	SharedKey   []byte
	Counter     uint16
	Tx          uint64
	Rx          uint64
	Connections ConnectionPool
}

func (s *Session) AddConnection(conn *net.Conn) {
	s.Connections = append(s.Connections, conn)
}
func (s *Session) RemoveConn(element *net.Conn) {
	// Iterate over the slice to find the element
	var temp ConnectionPool
	for i := 0; i < len(s.Connections); i++ {
		if s.Connections[i] != element {
			temp = append(temp, s.Connections[i])
		}
	}
	s.Connections = temp
	fmt.Println(element, s)

}

type Sessions map[uint16]Session // [SessionID]Session

func (s Sessions) FindUser(userID uint16) (uint16, bool) {
	for _, session := range s {
		if session.UserID == userID {
			return session.ID, true
		}
	}
	return 0, false
}
