package handshake

import (
	"fmt"
	"io"
)

// Handshake is a special message that a peer uses to identify itself
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// New creates a new Handshake with the standard ptsr
func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize serializes the handshake to a buffer
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], []byte(h.Pstr))
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

// Read parses a handshake from a stream
func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	lengthPstr := int(lengthBuf[0])

	if lengthPstr == 0 {
		return nil, fmt.Errorf("lengthPstr cannot be 0")
	}

	handshakeBuf := make([]byte, 48+lengthPstr)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[lengthPstr+8:lengthPstr+8+20])
	copy(peerID[:], handshakeBuf[lengthPstr+8+20:])

	h := Handshake{
		Pstr:     string(handshakeBuf[0:lengthPstr]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
