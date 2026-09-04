package protocol

import (
	"encoding/binary"
	"errors"
)

const ACK_PACKET_LEN int = 1

type AckPacket struct{}

func CreateAckPacket() Packet {
	return &AckPacket{}
}

func (e *AckPacket) GetType() uint8 {
	return TYPE_ACK
}

func (e *AckPacket) Header() []byte {
	return []byte{e.GetType()}
}

func (e *AckPacket) Serialize() []byte {
	message := make([]byte, 0, ACK_PACKET_LEN+LENGTH_BYTES)
	message = binary.BigEndian.AppendUint16(message, uint16(ACK_PACKET_LEN))
	message = append(message, e.Header()...)
	return message
}

func AckFromBytes(bytes []byte) (*AckPacket, error) {
	if len(bytes) != ACK_PACKET_LEN {
		return nil, errors.New("Invalid packet length")
	}
	if bytes[0] != TYPE_ACK {
		return nil, errors.New("Invalid packet type")
	}
	return &AckPacket{}, nil
}
