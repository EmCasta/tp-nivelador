import safe_socket
from protocol.packet import LENGTH_BYTES

def read_packet(socket):
    """
    Lee un paquete segun el protocolo y retorna sus bytes
    """
    # leer primero longitud
    message_length = safe_socket.recv_all(
        socket, LENGTH_BYTES
    )
    length = int.from_bytes(message_length, "big", signed=False)
    # leer paquete en si
    return safe_socket.recv_all(
        socket, length
    )
