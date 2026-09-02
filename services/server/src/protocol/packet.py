TYPE_HELLO = 0x00
TYPE_BET = 0x01
TYPE_END = 0x02
FIRST_BIT = 0b10000000
LAST_SEVEN_BITS = 0b01111111
BIT_OFFSET = 7
LENGTH_BYTES = 1

class Packet():

    def get_type(self) -> int:
        raise NotImplementedError("Subclase debe implementar get_type")

    def serialize(self) -> bytearray:
        raise NotImplementedError("Subclase debe implementar serialize")

    def header(self) -> bytearray:
        return self.get_type().to_bytes(1, "big", signed=False)

def set_last_packet_flag(packet: bytearray):
    first_byte = packet[LENGTH_BYTES]
    first_byte = first_byte | FIRST_BIT
    packet[LENGTH_BYTES] = first_byte

def get_last_packet_flag(packet: bytearray) -> bool:
    first_byte = packet[0]
    flag = (first_byte & FIRST_BIT) >> BIT_OFFSET
    is_last = flag == 1
    first_byte = first_byte & LAST_SEVEN_BITS
    packet[0] = first_byte
    return is_last
