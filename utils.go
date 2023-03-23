package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"net"
)

func generateRandom128BitString() string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(randomBytes)
}

func generateRandomPadding() []byte {
	// Generate a random integer between 1 and 1000
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(err)
	}
	randomLength := n.Int64() + 1 // Add 1 to ensure non-zero integer

	// Create a buffer to hold the random bytes
	randomBytes := make([]byte, randomLength)

	// Fill the buffer with random bytes
	_, err = rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return randomBytes
}

func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

// Convert a uint32 value to an IP address
func uint32ToIP(ipVal uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipVal)
	return ip
}
