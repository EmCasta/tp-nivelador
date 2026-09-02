from protocol.packet import Packet, TYPE_HELLO

HELLO_PACKET_LEN = 5

class HelloPacket(Packet):
    def __init__(self, agency_id: int):
        super().__init__()
        self.agency_id = agency_id

    def get_type(self):
        return TYPE_HELLO

    def serialize(self):
        message = bytes()
        message += self.header()
        message += self.agency_id.to_bytes(5, "big", signed=False)
        length = len(message).to_bytes(1, "big", signed=False)
        return length + message

def hello_packet_from_bytes(bytes: bytes) -> HelloPacket:
    if len(bytes) != HELLO_PACKET_LEN:
        raise ValueError("Invalid packet length")
    if bytes[0] != TYPE_HELLO:
        raise ValueError("Invalid packet type")

    agency_id = int.from_bytes(bytes[1:], "big", signed=False)
    return HelloPacket(agency_id)
