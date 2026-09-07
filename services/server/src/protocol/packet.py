TYPE_HELLO = 0x00
TYPE_BET = 0x01
TYPE_ACK = 0x02
FIRST_BIT = 0b10000000
LAST_SEVEN_BITS = 0b01111111
BIT_OFFSET = 7
LENGTH_BYTES = 2
ENDIANNESS = "big"
STR_ENCODING = "utf-8"

class Packet():
    """
    Clase abstracta Paquete, representa un paquete generico del protocolo
    """

    def get_type(self) -> int:
        """
        Permite obtener el tipo de paquete
        """
        raise NotImplementedError("Subclase debe implementar get_type")

    def serialize(self) -> bytearray:
        """
        Permite convertir el paquete a una tira de bytes
        """
        raise NotImplementedError("Subclase debe implementar serialize")

    def header(self) -> bytearray:
        """
        Retorna el header del paquete en bytes
        """
        return self.get_type().to_bytes(1, ENDIANNESS, signed=False)

def set_last_packet_flag(packet: bytearray, offset: int):
    """
    Setea el flag de ultimo paquete en el offset dado
    """
    first_byte = packet[offset]
    first_byte = first_byte | FIRST_BIT
    packet[offset] = first_byte

def get_last_packet_flag(packet: bytearray, offset: int) -> bool:
    """
    Obtiene el flag de ultimo paquete del offset dado
    """
    first_byte = packet[offset]
    flag = (first_byte & FIRST_BIT) >> BIT_OFFSET
    is_last = flag == 1
    first_byte = first_byte & LAST_SEVEN_BITS
    packet[offset] = first_byte
    return is_last
