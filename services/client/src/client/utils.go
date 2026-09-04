package client

import (
	"encoding/binary"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

func readPacket(conn io.Reader) ([]byte, error) {
	messageLength, err := safe_socket.RecvAll(conn, protocol.LENGTH_BYTES)
	if err != nil {
		return []byte{}, err
	}
	length := binary.BigEndian.Uint16(messageLength)
	return safe_socket.RecvAll(conn, int(length))
}
