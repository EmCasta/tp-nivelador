package client

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

const CSV_DELIMITER string = ","
const EXPECTED_FIELD_NUMBER int = 5
const BASE_10 int = 10
const BIT_SIZE int = 32
const MIN_LEN_PACKET int = 25

type Bet struct {
	AgencyId  uint32
	Document  uint32
	Number    uint32
	Birthdate string
	FirstName string
	LastName  string
}

// podria mover esto a un paquete storage o algo
func FromCsv(csv string, agencyId uint32) (Bet, error) {
	fields := strings.Split(csv, CSV_DELIMITER)
	if len(fields) != EXPECTED_FIELD_NUMBER {
		return Bet{}, errors.New("Unexpected number of fields")
	}

	firstName := fields[0]
	lastName := fields[1]
	birthdate := fields[3]
	document, err := strconv.ParseUint(fields[2], BASE_10, BIT_SIZE)
	if err != nil {
		return Bet{}, err
	}
	number, err := strconv.ParseUint(fields[4], BASE_10, BIT_SIZE)
	if err != nil {
		return Bet{}, err
	}

	bet := Bet{
		AgencyId:  agencyId,
		Document:  uint32(document),
		Number:    uint32(number),
		Birthdate: birthdate,
		FirstName: firstName,
		LastName:  lastName,
	}
	return bet, nil
}

func (b Bet) ToCsv() string {
	document := strconv.FormatUint(uint64(b.Document), BASE_10)
	number := strconv.FormatUint(uint64(b.Number), BASE_10)
	fields := []string{b.FirstName, b.LastName, document, b.Birthdate, number}
	return strings.Join(fields, CSV_DELIMITER)
}

func (b Bet) ToBytes(isLast bool) []byte {
	// primero byte con isLast (para que sea mas comodo luego agregar cant. de apuestas en batch)
	message := make([]byte, 0, MIN_LEN_PACKET)
	var lastByte uint8 = 0
	if isLast {
		lastByte = 1
	}
	message = append(message, lastByte)

	// luego agency id en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, b.AgencyId)

	// luego document en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, b.Document)

	// luego number en 32 bits unsigned, big endian
	message = binary.BigEndian.AppendUint32(message, b.Number)

	// luego birthdate, string de longitud 10
	message = append(message, []byte(b.Birthdate)...)

	firstName := []byte(b.FirstName)
	lastName := []byte(b.LastName)

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

func FromBytes(bytes []byte) (Bet, bool, error) {
	// parsear de bytes
	if len(bytes) < MIN_LEN_PACKET {
		return Bet{}, false, errors.New("Package too short")
	}

	// primer byte is Last (por ahora)
	isLast := bytes[0] == 1

	// agency id uint32, big endian
	agencyId := binary.BigEndian.Uint32(bytes[1:5])

	// luego document uint32, BE
	document := binary.BigEndian.Uint32(bytes[5:9])

	// luego number uint32, BE
	number := binary.BigEndian.Uint32(bytes[9:13])

	// luego birthdate, string len 10
	birthdate := string(bytes[13:23])
	// TODO: validar len 10 con el formato esperado!!

	// luego len de first name, uint8
	firstNameLen := uint8(bytes[23])

	// luego len de last name, uint8
	lastNameLen := uint8(bytes[24])

	// luego first name, string de largo variable
	if len(bytes) < MIN_LEN_PACKET+int(firstNameLen) {
		return Bet{}, false, errors.New("First Name too short")
	}

	firstName := string(bytes[MIN_LEN_PACKET : MIN_LEN_PACKET+int(firstNameLen)])

	// luego lastName, string de largo variable
	if len(bytes) < MIN_LEN_PACKET+int(firstNameLen)+int(lastNameLen) {
		return Bet{}, false, errors.New("Last Name too short")
	}

	lastName := string(bytes[MIN_LEN_PACKET+int(firstNameLen) : MIN_LEN_PACKET+int(firstNameLen)+int(lastNameLen)])

	bet := Bet{
		AgencyId:  agencyId,
		Document:  document,
		Number:    number,
		Birthdate: birthdate,
		FirstName: firstName,
		LastName:  lastName,
	}

	return bet, isLast, nil
}
