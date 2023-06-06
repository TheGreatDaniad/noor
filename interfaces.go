package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"

	"github.com/jackpal/gateway"
	"github.com/songgao/water"
)

type RoutingData struct {
	DefaultGateway net.IP
	serverAddress  net.IP
}

func findPhysicalInterface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagBroadcast != 0 {
			return &iface, nil
		}
	}

	return nil, errors.New("could not find network interface")
}
func createTunnelInterfaceClient(ip net.IP) (*water.Interface, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})

	if err != nil {
		panic(err)
	}
	defaultGatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		fmt.Printf("Failed to retrieve default gateway: %v\n", err)
		panic(err)
	}
	fmt.Println(defaultGatewayIP)
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sudo", "ip", "addr", "add", ip.String(), "dev", ifce.Name())
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Failed to configure tun interface: %v", err)
			return nil, err
		}

		cmd = exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1")
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Failed to enable IP forwarding: %v", err)
			return nil, err
		}
		cmd = exec.Command("ip", "route", "add", "default", "dev", ifce.Name())
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	case "darwin":

		// Configure the interface with an IP address and netmask
		cmd := exec.Command("sudo", "ifconfig", ifce.Name(), "inet", ip.String(), "10.0.10.1", "netmask", "255.255.255.0")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		cmd = exec.Command("sudo", "sysctl", "-w", "net.inet.ip.forwarding=1")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		cmd = exec.Command("sudo", "route", "change", "default", "-interface", ifce.Name())
		err = cmd.Run()

		CleanUpFunctions = append(CleanUpFunctions, func() {
			cmd = exec.Command("sudo", "route", "change", "default", defaultGatewayIP.String())
			err = cmd.Run()
		})
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		return ifce, nil
	case "windows":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return ifce, nil
}
func createTunnelInterfaceServer() (*water.Interface, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})

	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to enable IP forwarding: %v", err)
		return nil, err
	}
	cmd = exec.Command("sudo", "ip", "addr", "add", "10.0.10.1/24", "dev", ifce.Name())
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to configure network interface: %v", err)
		return nil, err
	}
	// Bring up the tunnel
	cmd = exec.Command("sudo", "ip", "link", "set", "dev", ifce.Name(), "up")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to bring up the network interface: %v", err)
		return nil, err
	}
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to set up NAT rule: %v", err)
	}
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", "eth0", "-o", "tun0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to execute command: %s", err)
	}
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", "tun0", "-o", "eth0", "-j", "ACCEPT")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to set up NAT rule: %v", err)
	}
	return ifce, nil
}
