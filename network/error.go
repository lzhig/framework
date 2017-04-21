package network

type ErrorInvalidPacketHeader struct {
	ErrorNetwork
}

type ErrorPacketSizeTooLarge struct {
	ErrorNetwork
}

type ErrorPacketBufferSizeTooSmall struct {
	ErrorNetwork
}

type ErrorNetwork struct {
	s string
	error
}

func (e *ErrorNetwork) Error() string {
	return e.s
}
