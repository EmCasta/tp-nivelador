package client

import (
	"encoding/binary"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

// Lee un paquete proveniente del servidor y lo parsea
func readPacket(conn io.Reader) ([]byte, error) {
	// leer bytes con longitud del paquete
	messageLength, err := safe_socket.RecvAll(conn, protocol.LENGTH_BYTES)
	if err != nil {
		return []byte{}, err
	}
	length := binary.BigEndian.Uint16(messageLength)
	// leer paquete en si
	return safe_socket.RecvAll(conn, int(length))
}
