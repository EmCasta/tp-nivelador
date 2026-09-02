package protocol

import (
	"encoding/binary"
	"errors"
)

const END_PACKET_LEN int = 1

type EndPacket struct{}

func CreateEndPacket() Packet {
	return &EndPacket{}
}

func (e *EndPacket) GetType() uint8 {
	return TYPE_END
}

func (e *EndPacket) Header() []byte {
	return []byte{e.GetType()}
}

func (e *EndPacket) Serialize() []byte {
	message := make([]byte, LENGTH_BYTES, END_PACKET_LEN+LENGTH_BYTES)
	message = binary.BigEndian.AppendUint16(message, uint16(END_PACKET_LEN))
	message = append(message, e.Header()...)
	return message
}

func EndFromBytes(bytes []byte) (*EndPacket, error) {
	if len(bytes) != END_PACKET_LEN {
		return nil, errors.New("Invalid packet length")
	}
	if bytes[0] != TYPE_END {
		return nil, errors.New("Invalid packet type")
	}
	return &EndPacket{}, nil
}
