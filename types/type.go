package types

import (
	"bytes"
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
