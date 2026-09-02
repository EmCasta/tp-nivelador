package protocol

import (
	"encoding/binary"
)

const HELLO_PACKET_LEN int = 6

type HelloPacket struct {
	agencyId  uint32
	batchSize uint8
}

func CreateHelloPacket(agencyId uint32, batchSize uint8) Packet {
	return &HelloPacket{agencyId, batchSize}
}

func (h *HelloPacket) GetType() uint8 {
	return TYPE_HELLO
}

func (h *HelloPacket) Header() []byte {
	return []byte{h.GetType()}
}

func (h *HelloPacket) Serialize() []byte {
	message := make([]byte, 0, HELLO_PACKET_LEN+LENGTH_BYTES)
	message = binary.BigEndian.AppendUint16(message, uint16(HELLO_PACKET_LEN))
	message = append(message, h.Header()...)

	message = binary.BigEndian.AppendUint32(message, h.agencyId)
	message = append(message, h.batchSize)
	return message
}
