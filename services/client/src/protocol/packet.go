package protocol

const (
	TYPE_HELLO      uint8 = 0x00
	TYPE_BET        uint8 = 0x01
	TYPE_END        uint8 = 0x02
	FIRST_BIT       byte  = 0b10000000
	LAST_SEVEN_BITS byte  = 0b01111111
	BIT_OFFSET      byte  = 7
	LENGTH_BYTES    int   = 1
)

type Packet interface {
	GetType() uint8
	Serialize() []byte
	Header() []byte
}

func SetLastPacketFlag(packet []byte) {
	firstByte := packet[0]
	firstByte = firstByte | FIRST_BIT
	packet[0] = firstByte
}

func GetLastPacketFlag(packet []byte) bool {
	firstByte := packet[0]
	flag := (firstByte & FIRST_BIT) >> BIT_OFFSET
	isLast := flag == 1
	firstByte = firstByte & LAST_SEVEN_BITS
	packet[0] = firstByte
	return isLast
}

// idea:
//	- quitar el end packet, que se mande un flag en el primer bit de tipo!
//	- cuando se terminan de procesar las apuestas de un batch, se manda el batch con el bit prendido
//	- el end packet lo puedo refactorizar como ack luego
//	- mandar hello: agency id + batch size LISTO
//	- recorrer apuestas de a n, ir poniendolas en el paquete.
//		- antes del paquete entero debe ir el tipo de paquete, no en cada bet LISTO
//		- antes del paquete entero debe ir la longitud del paquete entero LISTO
//		- no es necesario mandar longitud de cada bet por separado, le puedo pasar el offset al parser LISTO
