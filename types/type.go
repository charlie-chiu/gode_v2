package types

import (
	"bytes"
	"strconv"
)

/*
	uint8(byte)		represent unsigned integer 0~255
	uint16			0 ~ 65535
	uint32(rune)	0 ~ 4,294,967,295
*/

type GameType uint16

type GameCode uint16

type HallID uint16

type UserID uint32

type SessionID []byte

//todo: using pointer receiver instead
func (s SessionID) String() string {
	return string(s)
}

func (s *SessionID) UnmarshalJSON(b []byte) error {
	*s = bytes.Trim(b, `"`)

	return nil
}

type BetInfo []byte

//todo: using pointer receiver instead
func (i BetInfo) String() string {
	return string(i)
}

func (i *BetInfo) UnmarshalJSON(b []byte) error {
	*i = bytes.Trim(b, `"`)

	return nil
}

type Credit uint32

func (c *Credit) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, `"`)
	credit, err := strconv.ParseUint(string(b), 10, 0)
	if err != nil {
		return err
	}

	*c = Credit(credit)
	return nil
}
