package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/songgao/water"
	"github.com/spf13/viper"
	"golang.org/x/net/ipv4"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MaxSessions         uint32 `yaml:"max_sessions"`
	MaxSameUserSessions uint8  `yaml:"max_same_user_sessions"`
	Port                uint16 `yaml:"port"`
	BaseClientIP        net.IP `yaml:"base_client_ip"`
	ServerIP            net.IP `yaml:"server_ip"`
}
type Server struct {
	Config          Config
	Sessions        Sessions
	BaseLocalIP     net.IP
	TunnelInterface *water.Interface
}

func runServer() {
	config := readConfig()
	ip, err := findGlobalIP()
	if err != nil {
		panic("cannot find physical interface ip of the server")
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", ip.To4().String(), config.Port))
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}

	defer listener.Close()
	fmt.Println("listening at port:", config.Port)
	ifce, err := createTunnelInterfaceServer()
	if err != nil {
		panic(err)
	}
	var server Server = Server{
		Config:          config,
		Sessions:        make(Sessions),
		TunnelInterface: ifce,
		BaseLocalIP:     net.IPv4(10, 0, 10, 1),
	}
	go handleServerIncomingResponses(server)
	// Accept incoming connections and handle them
	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleTCPConnection(&conn, server)
	}

}

func handleTCPConnection(conn *net.Conn, server Server) {
	// Read data from client
	buffer := make([]byte, BUFFER_SIZE)
	n, err := (*conn).Read(buffer)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}
	data := buffer[:n]
	sessionID, err := handleHandshakeTCPServer(data, conn, server)
	if err != nil {
		(*conn).Close()
		return
	}
	session := server.Sessions[sessionID]
	(*conn).Write(session.LocalIp)
	log.Printf("%v: Connection established with the following session info %+v ", time.Now(), session)
	buf := make([]byte, BUFFER_SIZE)
	// var totalBytes float64
	// i := 0
	var packets [][]byte
	var i int
	for {
		n, _ := (*conn).Read(buf)
		fmt.Println(n)
		i++
		fmt.Println("read: ", i)
		if n == 0 {
			s := server.Sessions[sessionID]
			s.RemoveConn(conn)
			server.Sessions[sessionID] = s
			break
		}
		// if isICMPPacket(buf[:n]) {
		// 	i += n
		// }
		// totalBytes += (float64(n) / 1000)
		// fmt.Println("client: ", totalBytes)
		// pkt, err := decrypt(session.SharedKey, buf[:n])
		// if err != nil {
		// 	continue
		// }
		packets = extractIPPackets(buf[:n])
		for _, p := range packets {
			server.TunnelInterface.Write(p)
		}

	}
}

func readConfig() Config {
	var config Config = Config{
		MaxSessions:         1024,
		MaxSameUserSessions: 6,
		Port:                56000,
		BaseClientIP:        net.IPv4(10, 0, 10, 1),
	}
	// maybe in future use /etc for configs
	// viper.AddConfigPath("/etc/noor")
	viper.SetConfigFile(CONFIG_FILE_PATH)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n using default config", err)
		return config
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config file: %s\n", err)
		panic(err)
	}
	if config.ServerIP.Equal(nil) {
		ip, err := findGlobalIP()
		if err != nil {
			panic(err)
		}
		config.ServerIP = ip
	}
	return config
}

func addSession(c *net.Conn, server Server, userID uint16, sharedKey []byte) (uint16, error) {
	cc := *c
	remoteAddr := cc.RemoteAddr()
	ip, ok := remoteAddr.(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("remote address is not a TCP address: %s", remoteAddr)
	}
	count := len(server.Sessions) + 1
	localIP, err := AddToIP("10.0.10.1", uint32(count))
	if err != nil {
		return 0, err
	}
	sessionID, found := server.Sessions.FindUser(userID)
	if found {
		s := server.Sessions[sessionID]
		s.AddConnection(c)
		server.Sessions[sessionID] = s
	} else {
		server.Sessions[uint16(len(server.Sessions)+1)] = Session{
			ID:          uint16(len(server.Sessions) + 1),
			UserID:      userID,
			RealIp:      ip.IP,
			Connections: []*net.Conn{c},
			SharedKey:   sharedKey,
			LocalIp:     localIP,
		}
	}

	return uint16(len(server.Sessions)), nil
}

func setupServer() {
	// Prompt user for port number
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter port number (default 56000): ")
	portString, _ := reader.ReadString('\n')
	portString = strings.TrimSpace(portString)
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		fmt.Println("Invalid port number. Please try again.")
		setupServer()
		return
	}
	if port == 0 {
		port = 56000
	}

	// Prompt user for maximum number of connected users
	fmt.Print("Enter maximum number of connected users: ")
	maxSessionsString, _ := reader.ReadString('\n')
	maxSessionsString = strings.TrimSpace(maxSessionsString)
	maxSessions, err := strconv.ParseUint(maxSessionsString, 10, 32)
	if err != nil {
		fmt.Println("Invalid maximum number of connected users. Please try again.")
		setupServer()
		return
	}

	// Prompt user for maximum concurrent connections of the same user
	fmt.Print("Enter maximum concurrent connections of the same user: ")
	maxSameUserSessionsString, _ := reader.ReadString('\n')
	maxSameUserSessionsString = strings.TrimSpace(maxSameUserSessionsString)
	maxSameUserSessions, err := strconv.ParseUint(maxSameUserSessionsString, 10, 8)
	if err != nil {
		fmt.Println("Invalid maximum concurrent connections of the same user. Please try again.")
		setupServer()
		return
	}

	// Create Config struct and write to YAML file
	config := Config{
		MaxSessions:         uint32(maxSessions),
		MaxSameUserSessions: uint8(maxSameUserSessions),
		Port:                uint16(port),
	}
	data, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Println("Error writing configuration to file")
		return
	}

	err = os.WriteFile(CONFIG_FILE_PATH, data, 0644)
	if err != nil {
		fmt.Println("Error writing configuration to file")
		return
	}

	fmt.Println("Server configuration written to config.yaml")
}

// handles the reponses from internet that comes to the tunnel interface of the server
func handleServerIncomingResponses(server Server) {

	buffer := make([]byte, BUFFER_SIZE)
	// var totalBytes float64
	i := 0
	for {
		n, _ := server.TunnelInterface.Read(buffer)
		if isICMPPacket(buffer[:n]) {
			i++
		}
		// totalBytes += float64(n) / 1000
		// fmt.Println("internet: ", totalBytes)
		routeServerIncomingResponses(server, buffer[:n])
	}
}

var j int

func routeServerIncomingResponses(server Server, packet []byte) {
	ipHeader, err := ipv4.ParseHeader(packet)
	if err != nil {
		return
	}
	subnet := net.IPNet{IP: net.ParseIP("10.0.10.0"), Mask: net.CIDRMask(24, 32)}
	if subnet.Contains(ipHeader.Dst) {
		id, err := IPToID(ipHeader.Dst, server.BaseLocalIP)
		if err != nil {
			return
		}
		j++
		var conn net.Conn
		_, ok := server.Sessions[id]
		if ok {
			conn = *server.Sessions[id].Connections.RandomPick()
			n, _ := conn.Write(packet)
			fmt.Println("wrote: ", j, "bytes: ", n)

		}

		// encrypted, err := encrypt(server.Sessions[id].SharedKey, packet)

	} else {
		return
	}
}
func isICMPPacket(packet []byte) bool {
	if len(packet) < 20 {
		// ICMP header is at least 20 bytes long
		return false
	}
	// Extract the protocol field from the IP header
	protocol := packet[9]
	return protocol == 1
}
