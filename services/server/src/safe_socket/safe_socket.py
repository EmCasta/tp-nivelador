import socket

def recv_all(socket: socket.socket, size):
    bytesRead = 0
    total = bytes()
    while bytesRead < size:
        message = socket.recv(size - bytesRead)
        total += message
        bytesRead += len(message)
        if len(message) == 0:
            return total
    return total


def send_all(socket: socket.socket, bytes):
    bytesWritten = 0
    while bytesWritten < len(bytes):
        n = socket.send(bytes[bytesWritten:])
        bytesWritten += n
    return bytesWritten
