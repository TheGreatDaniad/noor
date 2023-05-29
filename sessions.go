package main

import "net"

type Session struct {
	ID        uint16
	UserID    uint16
	RealIp    net.IP
	LocalIp   net.IP
	SharedKey []byte
	Counter   uint16
	Tx        uint64
	Rx        uint64
	Conn      *net.Conn
}

type Sessions map[uint16]Session // [SessionID]Session
