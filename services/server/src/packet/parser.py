from packet.packet import Packet, TYPE_BET, TYPE_END
from packet.bet_info_packet import bet_info_from_bytes
from packet.end_packet import end_from_bytes

HEADER_LEN = 1

class Deserializer:
    def deserialize(self, bytes: bytes, agency_id: int) -> Packet:
        if len(bytes) < HEADER_LEN:
            raise ValueError("Invalid header length")
        if bytes[0] == TYPE_BET:
            return bet_info_from_bytes(bytes, agency_id)
        if bytes[0] ==  TYPE_END:
            return end_from_bytes(bytes)
        raise ValueError("Unknown packet type")
