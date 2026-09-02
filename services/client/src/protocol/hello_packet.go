package protocol

import (
	"encoding/binary"
)

const HELLO_PACKET_LEN int = 5

type HelloPacket struct {
	agencyId uint32
}

func CreateHelloPacket(agencyId uint32) Packet {
	return &HelloPacket{agencyId}
}

func (h *HelloPacket) GetType() uint8 {
	return TYPE_HELLO
}

func (h *HelloPacket) Header() []byte {
	return []byte{h.GetType()}
}

func (h *HelloPacket) Serialize() []byte {
	message := make([]byte, 1, HELLO_PACKET_LEN+1)
	message[0] = uint8(HELLO_PACKET_LEN)
	message = append(message, h.Header()...)

	message = binary.BigEndian.AppendUint32(message, h.agencyId)
	return message
}
