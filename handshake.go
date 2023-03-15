package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
)

const (
	SimpleHanshakeMethod = 0
	// in simple handshake each user has a uint16 id and a password in the database
	// since the password is known by the
)

type ServerHandshakeModes map[uint8]func(Server, User, net.Conn) error
type ClientHandshakeModes map[uint8]func(net.Conn, string) ([]byte, error)

var serverHandshakeHandlers = ServerHandshakeModes{
	0: simpleHandshakeServer,
}
var clientHandshakeHandlers = ClientHandshakeModes{
	0: simpleHandshakeClient,
}

func handleHandshakeTCPServer(data []byte, c net.Conn, server Server) error {
	userIDBytes := data[0:2]
	userID := uint16(userIDBytes[0])<<8 | uint16(userIDBytes[1])

	u, err := findUserById(userID)
	if err != nil {
		log.Printf("error on handshake with the user with address:%v\n and packet:%v\nerror:%v", c.RemoteAddr(), data[:3], err)
		return err
	}
	handshakeByte := uint8(data[2])

	if handshakeByte >= 0xc0 {
		// send a list of supported handshaking methods
	} else if handshakeByte >= 0x80 {
		// send prefered method
	} else if handshakeByte < 0x40 {

		err := serverHandshakeHandlers[handshakeByte](server, u, c)
		if err != nil {
			return err
		}

	} else {
		return errors.New("invalid handshake byte")
	}
	return errors.New("unknown error on handshake")
}
func handleHandshakeTCPClient(c net.Conn, userID [2]byte, mode uint8, password string) ([]byte, error) {

	if mode >= 0xc0 {
		// TODO
	} else if mode >= 0x80 {
		// TODO
	} else if mode < 0x40 {
		packet := append(userID[:], mode)
		packet = append(packet, generateRandomPadding()...) // add random padding to improve obfuscation
		c.Write(packet)
		key, err := clientHandshakeHandlers[mode](c, password)
		if err != nil {
			return nil, err
		}
		return key, nil

	} else {
		return []byte{}, nil
	}
	return []byte{}, errors.New("unknown error")

}
func simpleHandshakeClient(c net.Conn, password string) ([]byte, error) {
	challenge := make([]byte, 16)
	b, err := c.Read(challenge)
	if err != nil || b != 16 {
		return []byte{}, errors.New("error on reading server's challenge")
	}
	hash := HashSha256(password)

	encryptedChallenge, err := encrypt(hash[:32], string(challenge))

	if err != nil {
		return []byte{}, errors.New("error on encrypting the challenge")
	}
	fmt.Println(encryptedChallenge)
	c.Write([]byte(encryptedChallenge))

	publicKeyServerBytes := make([]byte, 16)

	b, err = c.Read(publicKeyServerBytes)
	if err != nil || b != 16 {
		return []byte{}, errors.New("error on reading server's response")
	}
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return []byte{}, err
	}
	publicKey := privateKey.PublicKey()

	c.Write(publicKey.Bytes())

	return publicKeyServerBytes, nil
}

func simpleHandshakeServer(server Server, u User, c net.Conn) error {
	challenge := generateRandom128BitString()
	hash := HashSha256(challenge)

	c.Write([]byte(challenge))
	challengeResponse := make([]byte, 200)
	_, err := c.Read([]byte(challengeResponse))

	if err != nil {
		return errors.New("error on reading client's response")

	}
	rawResponse, err := decrypt([]byte(hash[:32]), challengeResponse)
	fmt.Println(rawResponse)
	if err != nil {
		errors.New("error on decrypting the challenge response")
	}
	if challenge != rawResponse {
		errors.New("authentication failed, the response is wrong.")

	}
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	publicKey := privateKey.PublicKey()
	c.Write(publicKey.Bytes())
	publicKeyClientBytes := make([]byte, 32)
	b, err := c.Read(publicKeyClientBytes)
	fmt.Println(publicKeyClientBytes)
	if err != nil || b != 32 {
		return errors.New("error on reading client's response")
	}

	publicKeyClient, err := curve.NewPublicKey(publicKeyClientBytes)
	if err != nil {
		return err
	}

	sharedKey, err := privateKey.ECDH(publicKeyClient)
	if err != nil {
		return err
	}
	fmt.Println(sharedKey, "shared")
	err = addSession(c, server, u.UserID, sharedKey)
	if err != nil {
		return err
	}
	return nil

}
