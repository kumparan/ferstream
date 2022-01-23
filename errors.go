package ferstream

import "errors"

var (
	// ErrBadUnmarshalResult given when unmarshal result from a message's Data is not as intended
	ErrBadUnmarshalResult = errors.New("ferstreamErr: bad unmarshal result")
	// ErrCastingPayloadToStruct given when unmarshal result from a message's Data is not as intended
	ErrCastingPayloadToStruct = errors.New("ferstreamErr: failed to cast payload to specified struct")
	// ErrGiveUpProcessingMessagePayload given when message's payload(data) is already processed x times, but always failed
	ErrGiveUpProcessingMessagePayload = errors.New("ferstreamErr: give up processing message payload")
	// ErrNilMessagePayload given when message's payload(data) is nil
	ErrNilMessagePayload = errors.New("ferstreamErr: nil message payload given")
)
