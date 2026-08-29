package safe_socket

import (
	"io"
)

func SendAll(socket io.Writer, bytes []byte) error {
	bytesWritten := 0
	for bytesWritten < len(bytes) {
		n, err := socket.Write(bytes[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	total := make([]byte, 0, size)
	bytesRead := 0
	for bytesRead < size {
		buff := make([]byte, size-bytesRead)
		n, err := socket.Read(buff)
		total = append(total, buff[:n]...)
		bytesRead += n
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return nil, err
		}
	}
	return total, nil
}
