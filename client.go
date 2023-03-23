package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

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
	connectToServer(host, port, userID, password)
}

func connectToServer(address string, port string, userID [2]byte, password string) {

	conn, err := net.Dial("tcp", address+":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %s\n", err.Error())
		os.Exit(1)
	}

	handshakeMode := uint8(0x00) // hardcoded for now but later make it more sophisticated
	_, err = handleHandshakeTCPClient(conn, userID, handshakeMode, password)
	if err != nil {
		log.Println("handshake failed", err)
		conn.Close()
		return
	}

	ifce, err := createTunnelInterfaceClient()
	if err != nil {

		log.Panicln(err)
	}

	packetBuf := make([]byte, 1500)
	for {
		n, err := ifce.Read(packetBuf)
		if err != nil {
			fmt.Println("Error reading from tunnel interface:", err)
			return
		}
		_, err = conn.Write(packetBuf[:n])
		if err != nil {
			fmt.Println("Error sending packet to server:", err)
			return
		}
	}

}
