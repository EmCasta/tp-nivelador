package lottery

import (
	"errors"
	"strconv"
	"strings"
)

const CSV_DELIMITER string = ","
const EXPECTED_FIELD_NUMBER int = 5
const BASE_10 int = 10
const BIT_SIZE int = 32

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
