// Packet schema:

// +-------------------+
// | Mode Byte (1 byte)|
// +-------------------+
// | User Identifier   |
// | (2 bytes)         |
// +-------------------+
// | Payload           |
// | (Variable length) |
// +-------------------+

// In this packet schema, the first byte (Mode Byte) specifies
// the mode of the packet. The first 2 bits of the Mode Byte are
// used to identify the mode, while the rest of the bits are
// reserved and for now are randomly generated on the fly for the
// obfuscation purposes.

// If the mode is "Private", the next 2 bytes are the User
// Identifier which is a unique identifier for the user.
// The remaining bytes in the packet are the Payload,
// which can have a variable length.

// If the mode is "Public", the next 6 bytes are reserved
// for future use. The remaining bytes in the packet are
// the Payload, which can have a variable length.

// Please note that this is a high-level description of the
// packet schema, and some of the details may need to be
// refined as the protocol is further developed.

// Modes:
// 00 - data in private mode
// 01 - handshake in public mode
// 10 - data in public mode
// 11 - hanshake in public mode

// third byte determines how the rest of 5 bits should be treated,
// if it is 0, it means the rest of bits are random and should not be considered
// if it is 1, it means they have a meaning and should be considered for
// detecting the mode
// this is for future development in case the protocol needed more data in the
// mode byte

package main

import (
	"encoding/binary"
	"math/rand"
	"net"
	"sort"
	"time"

	"golang.org/x/net/ipv4"
)

func generateModeByte(private bool, handshake bool, extended bool) byte {
	var base byte = 0
	if private {
		base = base | 0b10000000
	}
	if handshake {
		base = base | 0b01000000
	}
	if extended {
		base = base | 0b00100000
	}
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(32)
	// Use bitwise OR to set the rest of the bits to the randomly generated number
	modeByte := base | byte(randomNumber)

	// Return the generated modeByte
	return modeByte
}

func generateHandshakeModeByte(handshakeMode uint8) byte {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(32)
	// Use bitwise OR to set the rest of the bits to the randomly generated number
	modeByte := handshakeMode | byte(randomNumber)<<3
	// Return the generated modeByte
	return modeByte
}

func generatePrivatePacket(modeByte byte, handshakeModeByte byte, userId [2]byte, payload []byte) []byte {
	packet := []byte{}
	packet = append(packet, modeByte)
	packet = append(packet, userId[:]...)
	packet = append(packet, modeByte)
	packet = append(packet, payload...)
	return packet
}

type Packet struct {
	Handshake     bool
	Public        bool
	HandshakeMode uint8
	UserID        uint16
	Payload       []byte
}

func UnmarshalPacket(p []byte) Packet {
	packet := Packet{}
	modeByte := p[0]
	if modeByte&0x80 == 1 { // is first leftmost bit one?
		packet.Public = false
		userID := binary.BigEndian.Uint16(p[1:3])
		HandshakeMode := uint8(p[3] & 0b00000111)
		packet.UserID = userID
		packet.HandshakeMode = HandshakeMode
		packet.Payload = p[4:]
	}
	if modeByte&0x40 == 1 { // is second leftmost bit one?
		packet.Handshake = true
	}
	return packet
}

func changePacketSrc(packet []byte, src net.IP) ([]byte, error) {
	ipHeader, err := ipv4.ParseHeader(packet)
	if err != nil {
		return []byte{}, err
	}
	ipHeader.Src = src

	payload := packet[ipHeader.Len:]
	newHeader, err := ipHeader.Marshal()
	if err != nil {
		return []byte{}, err
	}
	return append(newHeader, payload...), nil

}
func extractIPPackets(data []byte) [][]byte {
	var packets [][]byte
	if len(data) < 4 {
		return packets
	}
	for len(data) >= 20 {

		packetLength := int(data[2])<<8 + int(data[3])
		if len(data) < packetLength || packetLength < 20 {
			return packets
		}
		packetData := data[:packetLength]
		packets = append(packets, packetData)

		// Update the remaining data
		data = data[packetLength:]
	}

	return packets
}

type PacketConstructionList map[uint16]PacketFragments

type PacketFragments struct {
	Fragments          []Fragment
	LastInsertedDate   time.Time
	Completed          bool
	LastPacketInserted bool
}
type Fragment struct {
	Offset int
	Data   []byte
}

func (pf *PacketFragments) AddFragment(f Fragment) {
	pf.Fragments = append(pf.Fragments, f)
	pf.LastInsertedDate = time.Now()
}
func (pf *PacketFragments) AddLastFragment(f Fragment) {
	pf.Fragments = append(pf.Fragments, f)
	pf.LastInsertedDate = time.Now()

}
func assemblePacket(fragments []Fragment) []byte {
	// Sort the fragments based on their offset
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Offset < fragments[j].Offset
	})

	// Calculate the total length of the packet
	totalLength := 0
	for _, fragment := range fragments {
		totalLength += len(fragment.Data)
	}

	// Create a buffer to hold the reassembled packet
	packet := make([]byte, totalLength)

	// Copy the fragment data into the packet buffer at the correct offset
	for _, fragment := range fragments {
		copy(packet[fragment.Offset:], fragment.Data)
	}

	return packet
}

func (pfl PacketConstructionList) RemoveOldFragments() {
	for id, fragments := range pfl {
		if fragments.LastInsertedDate.Before(time.Now().Add(-10 * time.Second)) {
			delete(pfl, id)
		}
	}
}
