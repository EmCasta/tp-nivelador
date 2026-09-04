from protocol.packet import Packet, TYPE_ACK

ACK_PACKET_LEN = 1

class AckPacket(Packet):
    def __init__(self):
        super().__init__()

    def get_type(self):
        return TYPE_ACK

    def serialize(self):
        message = bytearray()
        length = ACK_PACKET_LEN.to_bytes(2, "big", signed=False)
        message.extend(length)
        message.extend(self.header())
        return message

def ack_from_bytes(bytes: bytes) -> AckPacket:
    if len(bytes) != ACK_PACKET_LEN:
        raise ValueError("Invalid packet length")
    if bytes[0] != TYPE_ACK:
        raise ValueError("Invalid packet type")
    return AckPacket()
