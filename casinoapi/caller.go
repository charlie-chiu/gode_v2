package casinoapi

type Caller interface {
	Call(service, functionName string, parameters ...interface{}) ([]byte, error)
}
