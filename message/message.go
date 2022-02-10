package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0 // MsgChoke chokes the receiver
	MsgUnchoke       messageID = 1 // MsgUnchoke unchokes the receiver
	MsgInterested    messageID = 2 // MsgInterested expresses interest in receiving data
	MsgNotInterested messageID = 3 // MsgNotInterested expresses disinterest in receiving data
	MsgHave          messageID = 4 // MsgHave alerts the receiver that the sender has downloaded a piece
	MsgBitfield      messageID = 5 // MsgBitfield encodes which pieces the sender has downloaded
	MsgRequest       messageID = 6 // MsgRequest requests a block of data from the receiver
	MsgPiece         messageID = 7 // MsgPiece delivers a block of data to fulfill a request
	MsgCancel        messageID = 8 // MsgCancel cancels a request
)

// Message stores the ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// NewRequest creates a REQUEST Message
func NewRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// NewHave creates a HAVE Message
func NewHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// ParsePiece parses a PIECE Message amd copies its payload in a buffer
func (msg *Message) ParsePiece(expectedIndex int, buf []byte) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short, %d < 8", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if index != expectedIndex {
		return 0, fmt.Errorf("expected index %d, got %d", expectedIndex, index)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset too high, %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

// ParseHave parses a HAVE Message
func (msg *Message) ParseHave() (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}
	length := uint32(len(msg.Payload) + 1) // +1 for id
	buf := make([]byte, length+4)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(msg.ID)
	copy(buf[5:], msg.Payload)
	return buf
}

// Read parses a message from a stream.
// Returns `nil` on keep-alive message.
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	msg := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &msg, nil
}

func (msg *Message) name() string {
	if msg == nil {
		return "KeepAlive"
	}
	switch msg.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", msg.ID)
	}
}

func (msg *Message) String() string {
	if msg == nil {
		return msg.name()
	}
	return fmt.Sprintf("%s [%d]", msg.name(), len(msg.Payload))
}
