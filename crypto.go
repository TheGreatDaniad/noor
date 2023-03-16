package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/sha3"
)

// Encrypt encrypts plaintext using AES-256 in CBC mode with a random IV.
func encrypt(key, plaintext []byte) ([]byte, error) {
	// Generate a random IV (initialization vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Create a new AES cipher using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Pad the plaintext to a multiple of the block size
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	// Create a new CBC mode cipher using the AES cipher and IV
	mode := cipher.NewCBCEncrypter(block, iv)

	// Encrypt the padded plaintext
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// Prepend the IV to the ciphertext
	ciphertext = append(iv, ciphertext...)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256 in CBC mode.
func decrypt(key, ciphertext []byte) ([]byte, error) {
	// Extract the IV from the ciphertext
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create a new AES cipher using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new CBC mode cipher using the AES cipher and IV
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Unpad the plaintext
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// pkcs7Pad pads data to a multiple of blockSize using the PKCS#7 padding scheme.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

// pkcs7Unpad removes PKCS#7 padding from data.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty input")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}
	for i := len(data) - 1; i >= len(data)-padding; i-- {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
func HashSha256(input string) string {
	// Define the input word as a string
	inputBytes := []byte(input)
	sha3Hash := sha3.New256()
	sha3Hash.Write(inputBytes)
	digest := sha3Hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(digest)

}
