package server

import (
	"airdispat.ch/errors"
	adMessage "airdispat.ch/message"
	"getmelange.com/zooko/message"
)

func (z *ZookoServer) handleMessage(data []byte, mesType string, h adMessage.Header) (adMessage.Message, error) {
	switch mesType {
	case message.TypeRequestFunds:
		return z.handleRequestFunds(data, h)
	case message.TypeLookupName:
		return z.handleLookupName(data, h)
	case message.TypeRegisterName:
		return z.handleRegisterName(data, h)
	case message.TypeRenewName:
		return z.handleRenewName(data, h)
	default:
		return errors.CreateError(
			errors.UnexpectedError,
			"Message is of an incorrect type.",
			z.Key.Address,
		), nil
	}
}
