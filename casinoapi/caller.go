package casinoapi

type Caller interface {
	Call(service, function string, parameters ...interface{}) ([]byte, error)
}
