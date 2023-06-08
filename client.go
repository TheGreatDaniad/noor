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
	"golang.org/x/net/ipv4"
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
	connectToServer(host, port, userID, password)
}

func connectToServer(address string, port string, userID [2]byte, password string) {

	conn, err := net.Dial("udp", address+":"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %s\n", err.Error())
		os.Exit(1)
	}

	// handshakeMode := uint8(0x00) // hardcoded for now but later make it more sophisticated
	// ip, key, err := handleHandshakeTCPClient(conn, userID, handshakeMode, password)
	// if err != nil {
	// 	log.Println("handshake failed", err)
	// 	conn.Close()
	// 	return
	// }
	buf := make([]byte, 1500)
	_, err = conn.Read(buf)
	fmt.Println(buf, err)
	// ifce, err := createTunnelInterfaceClient(ip)
	if err != nil {
		log.Panicln(err)
	}
	// go handleSendPackets(ifce, key, conn)
	// go handleReceivePackets(ifce, key, conn)
	for {
	}
}
func handleReceivePackets(ifce *water.Interface, key []byte, conn net.Conn) {

	packetBuf := make([]byte, BUFFER_SIZE)
	// c, err := net.ListenIP("ip4:tcp", nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	for {
		n, _ := conn.Read(packetBuf)
		// decrypted, err := decrypt(key, packetBuf[:n])
		// if err != nil {
		// 	fmt.Println("Failed to decrypt the packet:", err)
		// 	return
		// }

		ifce.Write(packetBuf[:n])

	}
}

func handleSendPackets(ifce *water.Interface, key []byte, conn net.Conn) {
	packetBuf := make([]byte, BUFFER_SIZE)
	var totalBytes int

	for {
		n, _ := ifce.Read(packetBuf)
		totalBytes += n
		fmt.Println(totalBytes / 1000)
		handleBuf := func() {
			ipHeader, err := ipv4.ParseHeader(packetBuf[:n])
			if err != nil {
				return
			}
			subnet := net.IPNet{IP: net.ParseIP("10.0.10.0"), Mask: net.CIDRMask(24, 32)}
			if subnet.Contains(ipHeader.Dst) {
				fmt.Println("containes")
				return
			} else {

				if err != nil {
					fmt.Println("Failed to encrypt the packet:", err)
					return
				}
				_, err = conn.Write(packetBuf[:n])
				if err != nil {
					fmt.Println("Error sending packet to server:", err)
					return
				}
			}
		}
		handleBuf()

	}
}

// CreatePingPacket creates an ICMP ping packet with a given identifier and sequence number
func CreatePingPacket(identifier, sequenceNum uint16) []byte {
	// Create the ICMP echo request packet
	icmpPacket := make([]byte, 8)

	icmpPacket[0] = 8 // Type icmp
	icmpPacket[1] = 0 // Code
	icmpPacket[2] = 0 // Checksum (zeroed for now)
	icmpPacket[3] = 0
	icmpPacket[4] = byte(identifier >> 8)    // Identifier (high byte)
	icmpPacket[5] = byte(identifier & 0xff)  // Identifier (low byte)
	icmpPacket[6] = byte(sequenceNum >> 8)   // Sequence number (high byte)
	icmpPacket[7] = byte(sequenceNum & 0xff) // Sequence number (low byte)

	checksum := calculateChecksum(icmpPacket)
	icmpPacket[2] = byte(checksum >> 8)   // Set the checksum (high byte)
	icmpPacket[3] = byte(checksum & 0xff) // Set the checksum (low byte)

	// Create the IP packet
	ipPacket := make([]byte, 20+len(icmpPacket))

	ipPacket[0] = 0x45                       // Version and Header Length
	ipPacket[1] = 0                          // TOS
	ipPacket[2] = byte(len(ipPacket) >> 8)   // Total Length (high byte)
	ipPacket[3] = byte(len(ipPacket) & 0xff) // Total Length (low byte)
	ipPacket[4] = 0                          // Identification (high byte)
	ipPacket[5] = 0                          // Identification (low byte)
	ipPacket[6] = 0x40                       // Flags and Fragment Offset
	ipPacket[7] = 0                          // Fragment Offset
	ipPacket[8] = 64                         // TTL (Time to Live)
	ipPacket[9] = 1                          // Protocol (ICMP)
	ipPacket[10] = 0                         // Checksum (high byte)
	ipPacket[11] = 0                         // Checksum (low byte)
	ipPacket[12] = 0                         // Source IP address (zeroed for now)
	ipPacket[13] = 0
	ipPacket[14] = 0
	ipPacket[15] = 0
	ipPacket[16] = 4 // Destination IP address (4.2.2.4)
	ipPacket[17] = 2
	ipPacket[18] = 2
	ipPacket[19] = 4

	copy(ipPacket[20:], icmpPacket)

	return ipPacket
}
func calculateChecksum(data []byte) uint16 {
	var sum uint32

	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i+1])<<8 | uint32(data[i])
	}

	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1])
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16

	return uint16(^sum)
}

func cleanup(fns CleanUpFuncs) {
	for _, fn := range fns {
		fn()
	}
}

type CleanUpFuncs []func()
