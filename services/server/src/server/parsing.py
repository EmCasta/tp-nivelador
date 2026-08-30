from src_frozen.lottery.bet import Bet

# TODO: renombrar este archivo cuando ya tenga el modelo mas definido

MIN_LEN_PACKAGE = 25

def bet_to_bytes(bet: Bet, is_last: bool) -> bytes:
    message = bytes()

    # isLast, 1 byte
    is_last_flag = 1 if is_last else 0
    message +=  is_last_flag.to_bytes(1, "big", signed=False)

    # agency id, 32 bits unsigned, BE
    message += bet.agency_id.to_bytes(4, "big", signed=False)

    # document, 32 bits unsigned, BE
    message += bet.document.to_bytes(4, "big", signed=False)

    # number, 32 bits unsigned, BE
    message += bet.number.to_bytes(4, "big", signed=False)

    # birthdate, ASCII len 10
    message += bet.birthdate.encode(encoding="utf-8", errors="replace")

    first_name_bytes = bet.first_name.encode(encoding="utf-8", errors="replace")
    last_name_bytes = bet.last_name.encode(encoding="utf-8", errors="replace")

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

def bytes_to_bet(bytes: bytes) -> tuple[Bet, bool]:
    if len(bytes) < MIN_LEN_PACKAGE:
        raise ValueError("Package too short")

    # primer byte is last (por ahora)
    is_last = int.from_bytes(bytes[0], "big", signed=False) == 1
    
    # agency id, 4 bytes BE
    agency_id = int.from_bytes(bytes[1:5], "big", signed=False)

    # document, 4 bytes BE
    document = int.from_bytes(bytes[5:9], "big", signed=False)

    # number, 4 bytes BE
    number = int.from_bytes(bytes[9:13], "big", signed=False)

    # birthdate, string de len 10
    birthdate = bytes[13:23].decode(encoding="utf-8", errors="replace")
    # TODO: validar formato de birthdate!!

    # len de first name, byte unsigned
    first_name_len = int.from_bytes(bytes[23], "big", signed=False)

    # len de last name, byte unsigned
    last_name_len = int.from_bytes(bytes[24], "big", signed=False)

    # first name, string len variable
    if len(bytes) < MIN_LEN_PACKAGE + first_name_len:
        raise ValueError("First Name too short")

    first_name = bytes[MIN_LEN_PACKAGE:MIN_LEN_PACKAGE+first_name_len].decode(encoding="utf-8", errors="replace")

    # last name, string len variable
    if len(bytes) < MIN_LEN_PACKAGE + first_name_len + last_name_len:
        raise ValueError("Last Name too short")

    last_name = bytes[MIN_LEN_PACKAGE+first_name_len:MIN_LEN_PACKAGE+first_name_len+last_name_len].decode(encoding="utf-8", errors="replace")

    bet = Bet(agency_id, first_name, last_name, document, birthdate, number)
    return bet, is_last
