from protocol.packet import Packet, TYPE_HELLO

HELLO_PACKET_LEN = 6

class HelloPacket(Packet):
    def __init__(self, agency_id: int, batch_size: int):
        super().__init__()
        self.agency_id = agency_id
        self.batch_size = batch_size

    def get_type(self):
        return TYPE_HELLO

    def serialize(self):
        message = bytearray([0])
        message.extend(self.header())
        message.extend(self.agency_id.to_bytes(5, "big", signed=False))
        message.extend(self.batch_size.to_bytes(1, "big", signed=False))
        length = len(message).to_bytes(1, "big", signed=False)
        message[0] = length
        return message

def hello_packet_from_bytes(bytes: bytes) -> HelloPacket:
    if len(bytes) != HELLO_PACKET_LEN:
        raise ValueError("Invalid packet length")
    if bytes[0] != TYPE_HELLO:
        raise ValueError("Invalid packet type")

    agency_id = int.from_bytes(bytes[1:5], "big", signed=False)
    batch_size = int.from_bytes(bytes[5:], "big", signed=False)
    return HelloPacket(agency_id, batch_size)
