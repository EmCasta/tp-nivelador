package packet

const (
	TYPE_HELLO uint8 = 0x00
	TYPE_BET   uint8 = 0x01
	TYPE_END   uint8 = 0x02
)

type Packet interface {
	GetType() uint8
	Serialize() []byte
	Header() []byte
}
