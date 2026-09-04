package protocol

const (
	TYPE_HELLO      uint8 = 0x00
	TYPE_BET        uint8 = 0x01
	TYPE_ACK        uint8 = 0x02
	FIRST_BIT       byte  = 0b10000000
	LAST_SEVEN_BITS byte  = 0b01111111
	BIT_OFFSET      byte  = 7
	LENGTH_BYTES    int   = 2
)

type Packet interface {
	GetType() uint8
	Serialize() []byte
	Header() []byte
}

func SetLastPacketFlag(packet []byte, offset int) {
	firstByte := packet[offset]
	firstByte = firstByte | FIRST_BIT
	packet[offset] = firstByte
}

func GetLastPacketFlag(packet []byte, offset int) bool {
	firstByte := packet[offset]
	flag := (firstByte & FIRST_BIT) >> BIT_OFFSET
	isLast := flag == 1
	firstByte = firstByte & LAST_SEVEN_BITS
	packet[offset] = firstByte
	return isLast
}
