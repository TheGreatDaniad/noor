package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"

	"github.com/jackpal/gateway"
	"github.com/songgao/water"
	wgtun "golang.zx2c4.com/wireguard/tun"
)

type RoutingData struct {
	DefaultGateway net.IP
	serverAddress  net.IP
}

type Ifce struct {
	io.ReadWriteCloser
	Dev wgtun.Device
}

func (i Ifce) Read(p []byte) (n int, err error) {
	var bufs [][]byte
	var final []byte
	var sizes []int

	// Initialize bufs and sizes with a single element
	bufs = append(bufs, p)
	sizes = append(sizes, len(p))
	// Call the original Read method
	packetsRead, err := i.Dev.Read(bufs, sizes, 0)
	// Update the total number of bytes read
	for _, size := range sizes[:packetsRead] {
		n += size

	}
	for _, pckt := range bufs {
		final = append(final, pckt...)
	}
	return n, err
}
func (i Ifce) Write(p []byte) (n int, err error) {
	return i.Dev.Write([][]byte{p}, 0)
}
func (i Ifce) Close() (err error) {
	return i.Dev.Close()
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
func createTunnelInterfaceClient(ip net.IP, host string) (io.ReadWriteCloser, error) {

	defaultGatewayIP, err := gateway.DiscoverGateway()
	if err != nil {
		fmt.Printf("Failed to retrieve default gateway: %v\n", err)
		panic(err)
	}

	switch runtime.GOOS {
	case "linux":
		ifce, err := water.New(water.Config{
			DeviceType: water.TUN,
		})

		if err != nil {
			panic(err)
		}
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
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
	case "darwin":
		ifce, err := water.New(water.Config{
			DeviceType: water.TUN,
		})

		if err != nil {
			panic(err)
		}
		base := net.IPv4(ip.To4()[0], ip.To4()[0], ip.To4()[0], 1)
		fmt.Println(base)
		// Configure the interface with an IP address and netmask
		cmd := exec.Command("sudo", "ifconfig", ifce.Name(), "inet", ip.String(), base.String(), "netmask", "255.255.255.0")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		cmd = exec.Command("sudo", "ifconfig", ifce.Name(), "mtu", fmt.Sprintf("%v", BUFFER_SIZE))
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
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		cmd = exec.Command("sudo", "route", "add", host, defaultGatewayIP.String())
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		CleanUpFunctions = append(CleanUpFunctions, func() {
			cmd = exec.Command("sudo", "route", "change", "default", defaultGatewayIP.String())
			err = cmd.Run()
			cmd = exec.Command("sudo", "route", "delete", host)
			err = cmd.Run()
		})
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		return ifce, nil
	case "windows":
		tun, err := wgtun.CreateTUN("tun", 1500)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		t := Ifce{Dev: tun}

		cmd := exec.Command("netsh", "interface", "ipv4", "set", "address", "name=\"tun\"", "static", ip.String(), "255.255.255.0")
		fmt.Println(cmd)
		go func() {
			cmd.Run()
			if err != nil {
				log.Fatal(err)

			}
		}()

		if err != nil {
			panic(err)
		}
		fmt.Println(defaultGatewayIP.String())
		cmd = exec.Command("route", "add", host, "mask", "255.255.255.255", defaultGatewayIP.String())
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		// fmt.Println(defaultGatewayIP.String())

		cmd = exec.Command("route", "change", "0.0.0.0", "mask", "0.0.0.0", ip.String())
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface", "tun", fmt.Sprintf("mtu=%d", BUFFER_SIZE), "store=persistent")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		CleanUpFunctions = append(CleanUpFunctions, func() {
			cmd = exec.Command("route", "change", "0.0.0.0", "mask", "0.0.0.0", defaultGatewayIP.String())
			err = cmd.Run()
			cmd = exec.Command("route", "delete", host)
			err = cmd.Run()
			tun.Close()
		})
		return t, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil, errors.New("os not supported")
}
func createTunnelInterfaceServer(server Server) (*water.Interface, error) {
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
	cmd = exec.Command("sudo", "ip", "addr", "add", fmt.Sprintf("%s/24", server.BaseLocalIP), "dev", ifce.Name())
	fmt.Println(cmd)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to configure network interface: %v", err)
		return nil, err
	}
	cmd = exec.Command("sudo", "ifconfig", ifce.Name(), "mtu", fmt.Sprint(BUFFER_SIZE))
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
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
