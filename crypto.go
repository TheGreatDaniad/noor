package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/sha3"
)

// Encrypt a message using AES-128 CBC mode with a given key
func encrypt(key, plaintext string) (string, error) {
	// Convert the key and plaintext strings to byte slices
	keyBytes := []byte(key)
	plaintextBytes := []byte(plaintext)

	// Generate a new 256-bit key using AES-256
	aesKey, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Generate a random 128-bit IV (initialization vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Pad the plaintext to a multiple of the block size (128 bits)
	plaintextBytes = pad(plaintextBytes, aes.BlockSize)

	// Create a new CBC mode cipher using the AES-256 key and IV
	aesCipher := cipher.NewCBCEncrypter(aesKey, iv)

	// Encrypt the padded plaintext using CBC mode
	ciphertext := make([]byte, len(plaintextBytes))
	aesCipher.CryptBlocks(ciphertext, plaintextBytes)

	// Prepend the IV to the ciphertext
	ciphertext = append(iv, ciphertext...)

	// Encode the ciphertext and IV as a hexadecimal string
	ciphertextString := hex.EncodeToString(ciphertext)

	// Return the encrypted ciphertext as a string
	return ciphertextString, nil
}
func decrypt(key, ciphertext []byte) (string, error) {

	// Extract the IV (initialization vector) from the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create a new CBC mode cipher using the AES-256 key and IV
	aesKey, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesCipher := cipher.NewCBCDecrypter(aesKey, iv)

	// Decrypt the ciphertext using CBC mode
	plaintext := make([]byte, len(ciphertext))
	aesCipher.CryptBlocks(plaintext, ciphertext)

	// Remove padding from the plaintext
	plaintext = unpad(plaintext)

	// Return the decrypted plaintext as a string
	return string(plaintext), nil
}

func HashSha256(input string) string {
	// Define the input word as a string
	inputBytes := []byte(input)
	sha3Hash := sha3.New256()
	sha3Hash.Write(inputBytes)
	digest := sha3Hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(digest)

}

// pad appends padding to the plaintext to ensure its length is a multiple of blockSize
func pad(plaintext []byte, blockSize int) []byte {
	padding := blockSize - (len(plaintext) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

// unpad removes padding from the plaintext
func unpad(plaintext []byte) []byte {
	padding := int(plaintext[len(plaintext)-1])
	return plaintext[:len(plaintext)-padding]
}
