package casinoapi

import "gode/types"

type Caller interface {
	Call(service types.GameType, function string, parameters ...interface{}) ([]byte, error)
}
