package packet

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

const MIN_LEN_PACKET int = 21
const BIRTHDATE_DELIMITER string = "-"
const BIRTHDATE_LEN int = 10
const BIRTHDATE_FIELDS int = 3
const BIRTHDATE_YEAR_DIGITS int = 4
const BIRTHDATE_DAY_MONTH_DIGITS int = 2

type BetInfoPacket struct {
	Bet lottery.Bet
}

func CreateBetInfoPacket(bet lottery.Bet) Packet {
	return &BetInfoPacket{bet}
}

func (b *BetInfoPacket) GetType() uint8 {
	return TYPE_BET
}

func (b *BetInfoPacket) Header() []byte {
	return []byte{byte(b.GetType())}
}

func (b *BetInfoPacket) Serialize() []byte {
	message := make([]byte, 1, MIN_LEN_PACKET+1)

	// primero header
	message = append(message, b.Header()...)

	// luego document en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, b.Bet.Document)

	// luego number en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, b.Bet.Number)

	// luego birthdate, string de longitud 10
	message = append(message, []byte(b.Bet.Birthdate)...)

	firstName := []byte(b.Bet.FirstName)
	lastName := []byte(b.Bet.LastName)

	// luego longitud de first name, uint8
	message = append(message, uint8(len(firstName)))

	// luego longitud de last name, uint8
	message = append(message, uint8(len(lastName)))

	// luego first name per se
	message = append(message, firstName...)

	// luego last name per se
	message = append(message, lastName...)

	totalLen := len(message)
	message[0] = uint8(totalLen)
	return message
}

// recibe bytes con header
func BetInfoFromBytes(bytes []byte, agencyId uint32) (*BetInfoPacket, error) {
	// parsear de bytes
	if len(bytes) < MIN_LEN_PACKET {
		return nil, errors.New("Packet too short")
	}
	if bytes[0] != TYPE_BET {
		return nil, errors.New("Invalid packet type")
	}

	// document uint32, BE
	document := binary.BigEndian.Uint32(bytes[1:5])

	// luego number uint32, BE
	number := binary.BigEndian.Uint32(bytes[5:9])

	// luego birthdate, string len 10
	birthdate := string(bytes[9:19])
	if !validateBirthdate(birthdate) {
		return nil, errors.New("Invalid Birthdate format")
	}

	// luego len de first name, uint8
	firstNameLen := int(uint8(bytes[19]))

	// luego len de last name, uint8
	lastNameLen := int(uint8(bytes[20]))

	// luego first name, string de largo variable
	if len(bytes) < MIN_LEN_PACKET+firstNameLen {
		return nil, errors.New("First Name too short")
	}

	firstName := string(bytes[MIN_LEN_PACKET : MIN_LEN_PACKET+firstNameLen])

	// luego lastName, string de largo variable
	if len(bytes) < MIN_LEN_PACKET+firstNameLen+lastNameLen {
		return nil, errors.New("Last Name too short")
	}

	lastName := string(bytes[MIN_LEN_PACKET+firstNameLen : MIN_LEN_PACKET+firstNameLen+lastNameLen])

	bet := lottery.Bet{
		AgencyId:  agencyId,
		Document:  document,
		Number:    number,
		Birthdate: birthdate,
		FirstName: firstName,
		LastName:  lastName,
	}

	betInfo := &BetInfoPacket{bet}

	return betInfo, nil
}

func validateBirthdate(birthdate string) bool {
	if len(birthdate) != BIRTHDATE_LEN {
		return false
	}
	fields := strings.Split(birthdate, BIRTHDATE_DELIMITER)
	if len(fields) != BIRTHDATE_FIELDS {
		return false
	}
	for _, f := range fields {
		if _, err := strconv.ParseUint(f, lottery.BASE_10, lottery.BIT_SIZE); err != nil {
			return false
		}
	}
	if len(fields[0]) != BIRTHDATE_YEAR_DIGITS || len(fields[1]) != BIRTHDATE_DAY_MONTH_DIGITS || len(fields[2]) != BIRTHDATE_DAY_MONTH_DIGITS {
		return false
	}
	return true
}
