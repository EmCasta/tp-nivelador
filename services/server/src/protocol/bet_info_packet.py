from lottery.bet import Bet
from protocol.packet import Packet, TYPE_BET, LENGTH_BYTES

MIN_LEN_PACKET = 20
HEADER_LEN = 1
BIRTHDATE_DELIMITER = "-"
BIRTHDATE_LEN = 10
BIRTHDATE_FIELDS = 3
BIRTHDATE_YEAR_DIGITS = 4
BIRTHDATE_DAY_MONTH_DIGITS = 2

class BetInfoPacket(Packet):
    def __init__(self, bets: list[Bet]):
        super().__init__()
        self.bets = bets

    def get_type(self):
        return TYPE_BET

    def serialize(self):
        message = bytearray()
        message.extend(self.header())
    
        for bet in self.bets:
            message.extend(self._serialize_bet(bet))

        length = len(message)
        final_message = bytearray(length.to_bytes(2, "big", signed=False))
        final_message.extend(message)
        return final_message


    def _serialize_bet(self, bet):
        message = bytearray()
        message.extend(bet.document.to_bytes(4, "big", signed=False))
        message.extend(bet.number.to_bytes(4, "big", signed=False))
        message.extend(bet.birthdate.encode(encoding="utf-8", errors="replace"))
        first_name_bytes = bet.first_name.encode(encoding="utf-8", errors="replace")
        last_name_bytes = bet.last_name.encode(encoding="utf-8", errors="replace")
        first_name_len = len(first_name_bytes)
        message.append(first_name_len)
        last_name_len = len(last_name_bytes)
        message.append(last_name_len)
        message.extend(first_name_bytes)
        message.extend(last_name_bytes)
        return message

def bet_info_from_bytes(bytes: bytes, agency_id: int, batch_size: int) -> BetInfoPacket:
    if len(bytes) < HEADER_LEN:
        raise ValueError("Header too short")
    if bytes[0] != TYPE_BET:
        raise ValueError("Invalid packet type")

    offset = HEADER_LEN
    bets = []
    for _ in range(batch_size):
        if len(bytes[offset:]) == 0:
            break
        bet, offset = bet_from_bytes(bytes, offset, agency_id)
        bets.append(bet)
    packet = BetInfoPacket(bets)
    return packet

def bet_from_bytes(bytes: bytearray, offset: int, agency_id: int) -> tuple[Bet, int]:
    bet_packet = bytes[offset:]
    if len(bet_packet) < MIN_LEN_PACKET:
        raise ValueError("Packet too short")
    document = int.from_bytes(bet_packet[0:4], "big", signed=False)
    number = int.from_bytes(bet_packet[4:8], "big", signed=False)
    birthdate = bet_packet[8:18].decode(encoding="utf-8", errors="replace")
    if not validate_birthdate(birthdate):
        raise ValueError("Invalid Birthdate format")
    first_name_len = int.from_bytes(bet_packet[18:19], "big", signed=False)
    last_name_len = int.from_bytes(bet_packet[19:20], "big", signed=False)
    if len(bet_packet) < MIN_LEN_PACKET + first_name_len:
        raise ValueError("First Name too short")
    first_name = bet_packet[MIN_LEN_PACKET:MIN_LEN_PACKET+first_name_len].decode(encoding="utf-8", errors="replace")
    if len(bet_packet) < MIN_LEN_PACKET + first_name_len + last_name_len:
        raise ValueError("Last Name too short")
    last_name = bet_packet[MIN_LEN_PACKET+first_name_len:MIN_LEN_PACKET+first_name_len+last_name_len].decode(encoding="utf-8", errors="replace")
    bet = Bet(agency_id, first_name, last_name, document, birthdate, number)
    return bet, offset + MIN_LEN_PACKET + first_name_len + last_name_len

def validate_birthdate(birthdate: str) -> bool:
    if len(birthdate) != BIRTHDATE_LEN:
        return False
    fields = birthdate.split(BIRTHDATE_DELIMITER)
    if len(fields) != BIRTHDATE_FIELDS:
        return False
    if not all(f.isdigit() for f in fields):
        return False
    if len(fields[0]) != BIRTHDATE_YEAR_DIGITS or len(fields[1]) != BIRTHDATE_DAY_MONTH_DIGITS or len(fields[2]) != BIRTHDATE_DAY_MONTH_DIGITS:
        return False
    return True
