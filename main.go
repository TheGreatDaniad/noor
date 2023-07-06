package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	modeFlag   string
	portFlag   int
	userFlag   string
	hostFlag   string
	passwdFlag string
)

func init() {
	flag.StringVar(&modeFlag, "mode", "", "Select mode (server, client, user)")
	flag.StringVar(&modeFlag, "m", "", "Select mode (server, client, user)")
	flag.IntVar(&portFlag, "port", 56000, "Port to use")
	flag.IntVar(&portFlag, "P", 56000, "Port to use")
	flag.StringVar(&userFlag, "user", "", "Username to use")
	flag.StringVar(&userFlag, "u", "", "Username to use")
	flag.StringVar(&hostFlag, "host", "", "Host to connect to")
	flag.StringVar(&hostFlag, "h", "", "Host to connect to")
	flag.StringVar(&passwdFlag, "password", "", "Password to use")
	flag.StringVar(&passwdFlag, "p", "", "Password to use")
}

func main() {

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)
	defer cleanup(CleanUpFunctions)
	go func() {
		sig := <-signalCh
		log.Printf("Received signal: %v", sig)
		cleanup(CleanUpFunctions)
		os.Exit(0)
	}()
	flag.Parse()

	// Determine which function to run based on mode flag
	switch modeFlag {
	case "server", "s":
		runServer()
	case "client", "c":
		runClient(hostFlag, fmt.Sprintf("%v", portFlag), userFlag, passwdFlag) //TODO fix flag here
	case "user", "m":
		RunUserManager()
	default:
		var mode int

		// Prompt user for mode (server, client, or user management)
		for {
			fmt.Print("Select mode:\n1-Setup Server\n2-Run Server\n3-User Manager\n4-Run Client ")
			_, err := fmt.Scanln(&mode)
			if err != nil {
				fmt.Println("Invalid input. Please try again.")
				continue
			}
			if mode < 1 || mode > 4 {
				fmt.Println("Invalid mode selected. Please try again.")
				continue
			}
			break
		}

		// Determine which function to run based on mode
		switch mode {
		case 1:
			setupServer()
		case 2:
			runServer()
		case 3:
			RunUserManager()
		case 4:
			runClient("", "", "", "")
		default:
			fmt.Println("Invalid mode selected")
		}
	}
}
