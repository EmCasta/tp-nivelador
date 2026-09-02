from protocol.packet import Packet, TYPE_END

END_PACKET_LEN = 1

class EndPacket(Packet):
    def __init__(self):
        super().__init__()

    def get_type(self):
        return TYPE_END

    def serialize(self):
        message = bytes()
        message += self.header()
        length = len(message).to_bytes(1, "big", signed=False)
        return length + message

def end_from_bytes(bytes: bytes) -> EndPacket:
    if len(bytes) != END_PACKET_LEN:
        raise ValueError("Invalid packet length")
    if bytes[0] != TYPE_END:
        raise ValueError("Invalid packet type")
    return EndPacket()
