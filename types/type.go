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

func (i *HallID) UnmarshalJSON(b []byte) error {
	hid, err := trimToUint(b)
	if err != nil {
		return err
	}

	*i = HallID(hid)
	return nil
}

type UserID uint32

func (u *UserID) UnmarshalJSON(b []byte) error {
	uid, err := trimToUint(b)
	if err != nil {
		return err
	}

	*u = UserID(uid)
	return nil
}

type SessionID []byte

func (s SessionID) String() string {
	return string(s)
}

func (s *SessionID) UnmarshalJSON(b []byte) error {
	*s = bytes.Trim(b, `"`)

	return nil
}

type BetInfo []byte

func (i BetInfo) String() string {
	return string(i)
}

func (i *BetInfo) UnmarshalJSON(b []byte) error {
	*i = bytes.Trim(b, `"`)

	return nil
}

type Credit uint32

func (c *Credit) UnmarshalJSON(b []byte) error {
	credit, err := trimToUint(b)
	if err != nil {
		return err
	}

	*c = Credit(credit)
	return nil
}

func trimToUint(b []byte) (uint64, error) {
	b = bytes.Trim(b, `"`)
	uid, err := strconv.ParseUint(string(b), 10, 0)
	if err != nil {
		return 0, err
	}
	return uid, nil
}
