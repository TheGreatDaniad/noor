package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/songgao/water"
)

var CleanUpFunctions CleanUpFuncs

func runClient(host string, port string, userIDStr string, password string) {

	reader := bufio.NewReader(os.Stdin)
	if host == "" {
		fmt.Print("Enter server address: ")
		serverAddress, _ := reader.ReadString('\n')
		serverAddress = strings.TrimSpace(serverAddress)

	}
	if port == "" {
		fmt.Print("Enter server port: ")
		port, _ = reader.ReadString('\n')
		port = strings.TrimSpace(port)

	}
	if userIDStr == "" {
		fmt.Print("Enter user ID: ")
		userIDStr, _ = reader.ReadString('\n')
		userIDStr = strings.TrimSpace(userIDStr)

	}
	if password == "" {
		fmt.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)
	}
	n, err := strconv.ParseUint(userIDStr, 10, 16)
	if err != nil {
		fmt.Println(err)
		return
	}
	userID := [2]byte{byte(n >> 8), byte(n)}
	conns, key, ip, err := connectToServer(host, port, userID, password, 1)
	if err != nil {
		panic(err)
	}
	ifce, err := createTunnelInterfaceClient(ip, host)
	if err != nil {
		log.Panicln(err)
	}
	for _, c := range conns {
		go handleReceivePackets(ifce, key, *c)

	}
	handleSendPackets(ifce, key, conns)

}

func connectToServer(address string, port string, userID [2]byte, password string, connectionCount int) ([]*net.Conn, []byte, net.IP, error) {
	var conns []*net.Conn
	var ip net.IP
	var key []byte
	for i := 0; i < connectionCount; i++ {
		conn, err := net.Dial("tcp", address+":"+port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to server: %s\n", err.Error())
			os.Exit(1)
		}

		handshakeMode := uint8(0x00) // hardcoded for now but later make it more sophisticated
		ip, key, err = handleHandshakeTCPClient(conn, userID, handshakeMode, password)
		if err != nil {
			log.Println("handshake failed", err)
			conn.Close()
			for _, c := range conns {
				(*c).Close()
			}
			return nil, []byte{}, nil, err
		} else {

			conns = append(conns, &conn)
		}
	}
	return conns, key, ip, nil

}
func handleReceivePackets(ifce *water.Interface, key []byte, conn net.Conn) {

	packetBuf := make([]byte, BUFFER_SIZE)

	// var totalBytes float64
	i := 0

	for {
		i++
		n, _ := conn.Read(packetBuf)
		if n == 0 {
			conn.Close()
			fmt.Println("server closed the tcp connection")
			break
		}
		ifce.Write(packetBuf[:n])

	}
}

func handleSendPackets(ifce *water.Interface, key []byte, conns ConnectionPool) {
	packetBuf := make([]byte, BUFFER_SIZE)
	i := 0
	for {
		i++
		ifce.Read(packetBuf)

		(*conns.RandomPick()).Write(packetBuf)

	}
}

func cleanup(fns CleanUpFuncs) {
	for _, fn := range fns {
		fn()
	}
}

type CleanUpFuncs []func()
