package packet

import "errors"

const HEADER_LEN int = 1

type Deserializer interface {
	Deserialize([]byte, uint32) (Packet, error)
}

type Parser struct{}

func (p Parser) Deserialize(bytes []byte, agencyId uint32) (Packet, error) {
	if len(bytes) < HEADER_LEN {
		return nil, errors.New("Invalid header length")
	}
	switch bytes[0] {
	case TYPE_BET:
		return BetInfoFromBytes(bytes, agencyId)
	case TYPE_END:
		return EndFromBytes(bytes)
	default:
		return nil, errors.New("Unknown packet type")
	}
}
