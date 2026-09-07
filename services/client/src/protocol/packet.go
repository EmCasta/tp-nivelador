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

// Paquete generico que puede enviar/recibir el cliente
type Packet interface {
	// Obtiene el tipo del paquete
	GetType() uint8
	// Convierte el paquete a una serie de bytes
	Serialize() []byte
	// Retorna el header del paquete en bytes
	Header() []byte
}

// Permite obtener el header de un paquete
func GetPacketHeader(packet Packet) []byte {
	return []byte{packet.GetType()}
}

// Setea el bit de ultimo paquete al paquete recibido, en el offset dado
func SetLastPacketFlag(packet []byte, offset int) {
	firstByte := packet[offset]
	firstByte = firstByte | FIRST_BIT
	packet[offset] = firstByte
}

// Obtiene el flag de ultimo paquete, en el offset dado
func GetLastPacketFlag(packet []byte, offset int) bool {
	firstByte := packet[offset]
	flag := (firstByte & FIRST_BIT) >> BIT_OFFSET
	isLast := flag == 1
	firstByte = firstByte & LAST_SEVEN_BITS
	packet[offset] = firstByte
	return isLast
}
