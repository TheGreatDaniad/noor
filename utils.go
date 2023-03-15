package main

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
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
