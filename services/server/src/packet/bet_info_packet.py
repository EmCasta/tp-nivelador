from src_frozen.lottery.bet import Bet
from packet.packet import Packet, TYPE_BET

MIN_LEN_PACKET = 21
BIRTHDATE_DELIMITER = "-"
BIRTHDATE_LEN = 10
BIRTHDATE_FIELDS = 3
BIRTHDATE_YEAR_DIGITS = 4
BIRTHDATE_DAY_MONTH_DIGITS = 2

class BetInfoPacket(Packet):
    def __init__(self, bet: Bet):
        super().__init__()
        self.bet = bet

    def get_type(self):
        return TYPE_BET

    def serialize(self):
        message = bytes()

        message += self.header()
    
        # document, 32 bits unsigned, BE
        message += self.bet.document.to_bytes(4, "big", signed=False)
    
        # number, 32 bits unsigned, BE
        message += self.bet.number.to_bytes(4, "big", signed=False)
    
        # birthdate, utf8 len 10
        message += self.bet.birthdate.encode(encoding="utf-8", errors="replace")
    
        first_name_bytes = self.bet.first_name.encode(encoding="utf-8", errors="replace")
        last_name_bytes = self.bet.last_name.encode(encoding="utf-8", errors="replace")
    
        # longitud de first name, unsigned byte
        first_name_len = len(first_name_bytes)
        message += first_name_len.to_bytes(1, "big", signed=False)
    
        # longitud de last name, unsigned byte
        last_name_len = len(last_name_bytes)
        message += last_name_len.to_bytes(1, "big", signed=False)
    
        # first name per se
        message += first_name_bytes
    
        # last name per se
        message += last_name_bytes
        return message

def bet_info_from_bytes(bytes: bytes, agency_id: int) -> BetInfoPacket:
    if len(bytes) < MIN_LEN_PACKET:
        raise ValueError("Packet too short")
    if bytes[0] != TYPE_BET:
        raise ValueError("Invalid packet type")

    # document, 4 bytes BE
    document = int.from_bytes(bytes[1:5], "big", signed=False)

    # number, 4 bytes BE
    number = int.from_bytes(bytes[5:9], "big", signed=False)

    # birthdate, string de len 10
    birthdate = bytes[9:19].decode(encoding="utf-8", errors="replace")
    if not validate_birthdate(birthdate):
        raise ValueError("Invalid Birthdate format")

    # len de first name, byte unsigned
    first_name_len = int.from_bytes(bytes[19], "big", signed=False)

    # len de last name, byte unsigned
    last_name_len = int.from_bytes(bytes[20], "big", signed=False)

    # first name, string len variable
    if len(bytes) < MIN_LEN_PACKET + first_name_len:
        raise ValueError("First Name too short")

    first_name = bytes[MIN_LEN_PACKET:MIN_LEN_PACKET+first_name_len].decode(encoding="utf-8", errors="replace")

    # last name, string len variable
    if len(bytes) < MIN_LEN_PACKET + first_name_len + last_name_len:
        raise ValueError("Last Name too short")

    last_name = bytes[MIN_LEN_PACKET+first_name_len:MIN_LEN_PACKET+first_name_len+last_name_len].decode(encoding="utf-8", errors="replace")

    bet = Bet(agency_id, first_name, last_name, document, birthdate, number)
    return BetInfoPacket(bet)

def validate_birthdate(birthdate: str) -> bool:
    if len(birthdate) != BIRTHDATE_LEN:
        return False
    fields = birthdate.split(BIRTHDATE_DELIMITER)
    if len(fields) != BIRTHDATE_FIELDS:
        return False
    if not all(f.is_digit() for f in fields):
        return False
    if len(fields[0]) != BIRTHDATE_YEAR_DIGITS or len(fields[1]) != BIRTHDATE_DAY_MONTH_DIGITS or len(fields[2]) != BIRTHDATE_DAY_MONTH_DIGITS:
        return False
    return True
