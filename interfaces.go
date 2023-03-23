package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"

	"github.com/songgao/water"
)

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
func createTunnelInterfaceClient() (*water.Interface, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatal(err)
	}
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sudo", "ip", "addr", "add", "10.0.0.1/24", "dev", ifce.Name())
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
		cmd := exec.Command("sudo", "ifconfig", ifce.Name(), "inet", "10.0.10.1", "10.0.10.1", "netmask", "255.255.255.0")
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

		cmd = exec.Command("sudo", "route", "-n", "add", "-net", "0/1", "10.0.10.1")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		physicalInterface, err := findPhysicalInterface()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		cmd = exec.Command("sudo", "route", "-n", "add", "-net", "10.0.10.1", "-interface", physicalInterface.Name)
		err = cmd.Run()
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
	cmd := exec.Command("sudo", "ip", "addr", "add", "10.0.10.1/24", "dev", ifce.Name())
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to configure network interface: %v", err)
		return nil, err
	}
	return nil, nil
}
