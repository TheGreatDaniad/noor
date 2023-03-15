package main

import "net"

type Session struct {
	ID        uint32
	UserID    uint16
	RealIp    net.IP
	SharedKey []byte
	Counter   uint32
	Tx        uint64
	Rx        uint64
	Conn      *net.Conn
}

type Sessions map[uint32]Session // [SessionID]Session
