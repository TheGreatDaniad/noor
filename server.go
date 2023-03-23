package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

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
	Config         Config
	Sessions       Sessions
	SessionCounter uint16
	IPCounter      uint16
}

func runServer() {
	config := readConfig()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", config.Port))
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("listening at port:", config.Port)
	var server Server = Server{
		Config:         config,
		SessionCounter: 0,
		Sessions:       make(Sessions),
	}

	// Accept incoming connections and handle them
	for {
		conn, err := listener.Accept()
		conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // set default timeout to 30 seconds
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleTCPConnection(conn, server)
	}

}

func handleTCPConnection(conn net.Conn, server Server) {
	// Read data from client
	buffer := make([]byte, 1500)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}
	data := buffer[:n]
	err = handleHandshakeTCPServer(data, conn, server)
	if err != nil {
		conn.Close()
		return
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
	fmt.Println(config)
	return config
}

func addSession(c net.Conn, server Server, userID uint16, sharedKey []byte) error {
	remoteAddr := c.RemoteAddr()
	ip, ok := remoteAddr.(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("remote address is not a TCP address: %s", remoteAddr)
	}

	server.Sessions[server.SessionCounter] = Session{
		ID:        server.SessionCounter,
		UserID:    userID,
		RealIp:    ip.IP,
		Conn:      &c,
		SharedKey: sharedKey,
	}
	server.SessionCounter += 1
	return nil
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

func sendPacketsToInternet(packet []byte, socket int) {
	
	ipHeader, err := ipv4.ParseHeader(packet)
	if err != nil {
		fmt.Println("Failed to parse packet:", err)
		return
	}
	var dst [4]byte
	copy(dst[:], ipHeader.Dst.To4())
	sockaddr := syscall.SockaddrInet4{
		Port: 0,
		Addr: dst,
	}
	_, _, errno := syscall.Syscall6(
		syscall.SYS_WRITE,
		uintptr(socket),
		uintptr(unsafe.Pointer(&packet[0])),
		uintptr(len(packet)),
		uintptr(0),
		uintptr(unsafe.Pointer(&sockaddr)),
		uintptr(unsafe.Sizeof(sockaddr)),
	)
	fmt.Println("sent to the internet: ", ipHeader)
	println(errno)
}

func findGlobalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	// Iterate over interfaces
	for _, iface := range ifaces {
		// Check if interface is up and not a loopback or tunnel interface
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagPointToPoint == 0 {
			// Get list of addresses for interface
			addrs, err := iface.Addrs()
			if err != nil {
				panic(err)
			}

			// Iterate over addresses
			for _, addr := range addrs {
				// Check if address is an IPv4 or IPv6 global unicast address
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && !ip.IsLoopback() && ip.To4() != nil && ip.IsGlobalUnicast() {
					fmt.Println("Global IP address:", ip)
					return ip, nil
				}
			}
		}
	}

	return net.IP{}, nil

}
