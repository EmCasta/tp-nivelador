import safe_socket
from protocol.packet import LENGTH_BYTES

def read_packet(socket):
    message_length = safe_socket.recv_all(
        socket, LENGTH_BYTES
    )
    length = int.from_bytes(message_length, "big", signed=False)
    return safe_socket.recv_all(
        socket, length
    )
