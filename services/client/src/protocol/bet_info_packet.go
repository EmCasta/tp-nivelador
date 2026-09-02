package protocol

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

// TODO: mejorar limites de tamaño en bets!

const MIN_LEN_PACKET int = 20
const HEADER_LEN int = 1
const BIRTHDATE_DELIMITER string = "-"
const BIRTHDATE_LEN int = 10
const BIRTHDATE_FIELDS int = 3
const BIRTHDATE_YEAR_DIGITS int = 4
const BIRTHDATE_DAY_MONTH_DIGITS int = 2

type BetInfoPacket struct {
	Bets      []lottery.Bet
	batchSize int
}

func CreateBetInfoPacket(bets []lottery.Bet, batchSize int) Packet {
	return &BetInfoPacket{bets, batchSize}
}

func (b *BetInfoPacket) GetType() uint8 {
	return TYPE_BET
}

func (b *BetInfoPacket) Header() []byte {
	return []byte{byte(b.GetType())}
}

func (b *BetInfoPacket) Serialize() []byte {
	message := make([]byte, LENGTH_BYTES, b.batchSize*MIN_LEN_PACKET+HEADER_LEN+LENGTH_BYTES)
	message = append(message, b.Header()...)

	for _, bet := range b.Bets {
		message = append(message, b.serializeBet(bet)...)
	}
	totalLen := len(message)
	binary.BigEndian.PutUint16(message, uint16(totalLen-LENGTH_BYTES))
	return message
}

func (b *BetInfoPacket) serializeBet(bet lottery.Bet) []byte {
	message := make([]byte, 0, MIN_LEN_PACKET)

	// document en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, bet.Document)

	// luego number en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, bet.Number)

	// luego birthdate, string de longitud 10
	message = append(message, []byte(bet.Birthdate)...)

	firstName := []byte(bet.FirstName)
	lastName := []byte(bet.LastName)

	// luego longitud de first name, uint8
	message = append(message, uint8(len(firstName)))

	// luego longitud de last name, uint8
	message = append(message, uint8(len(lastName)))

	// luego first name per se
	message = append(message, firstName...)

	// luego last name per se
	message = append(message, lastName...)
	return message
}

func BetInfoFromBytes(bytes []byte, agencyId uint32, batchSize int) (*BetInfoPacket, error) {
	if len(bytes) < HEADER_LEN {
		return nil, errors.New("Header too short")
	}
	if bytes[0] != TYPE_BET {
		return nil, errors.New("Invalid packet type")
	}
	offset := HEADER_LEN
	bets := make([]lottery.Bet, 0, batchSize)
	for range batchSize {
		if len(bytes[offset:]) == 0 {
			break
		}
		bet, newOffset, err := betFromBytes(bytes, offset, agencyId)
		if err != nil {
			return nil, err
		}
		bets = append(bets, bet)
		offset = newOffset
	}
	packet := &BetInfoPacket{
		Bets:      bets,
		batchSize: batchSize,
	}
	return packet, nil
}

func betFromBytes(bytes []byte, offset int, agencyId uint32) (lottery.Bet, int, error) {
	// parsear de bytes
	betPacket := bytes[offset:]
	if len(betPacket) < MIN_LEN_PACKET {
		return lottery.Bet{}, 0, errors.New("Packet too short")
	}

	// document uint32, BE
	document := binary.BigEndian.Uint32(betPacket[0:4])

	// luego number uint32, BE
	number := binary.BigEndian.Uint32(betPacket[4:8])

	// luego birthdate, string len 10
	birthdate := string(betPacket[8:18])
	if !validateBirthdate(birthdate) {
		return lottery.Bet{}, 0, errors.New("Invalid Birthdate format")
	}

	// luego len de first name, uint8
	firstNameLen := int(uint8(betPacket[18]))

	// luego len de last name, uint8
	lastNameLen := int(uint8(betPacket[19]))

	// luego first name, string de largo variable
	if len(betPacket) < MIN_LEN_PACKET+firstNameLen {
		return lottery.Bet{}, 0, errors.New("First Name too short")
	}

	firstName := string(betPacket[MIN_LEN_PACKET : MIN_LEN_PACKET+firstNameLen])

	// luego lastName, string de largo variable
	if len(betPacket) < MIN_LEN_PACKET+firstNameLen+lastNameLen {
		return lottery.Bet{}, 0, errors.New("Last Name too short")
	}

	lastName := string(betPacket[MIN_LEN_PACKET+firstNameLen : MIN_LEN_PACKET+firstNameLen+lastNameLen])

	bet := lottery.Bet{
		AgencyId:  agencyId,
		Document:  document,
		Number:    number,
		Birthdate: birthdate,
		FirstName: firstName,
		LastName:  lastName,
	}

	return bet, offset + MIN_LEN_PACKET + firstNameLen + lastNameLen, nil
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
