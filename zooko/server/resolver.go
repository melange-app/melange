package server

import (
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
	zMessage "getmelange.com/zooko/message"
)

func (r *ZookoServer) handleRequestFunds(data []byte, h message.Header) (message.Message, error) {
	msg := new(zMessage.RequestFunds)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	tx, err := r.Endow(*msg.Address)
	if err != nil {
		return nil, err
	}

	// Return the message that contains the data.
	return zMessage.CreateMessage(&zMessage.TransferFunds{
		Id:     &tx.TxID,
		Index:  &tx.Output,
		Amount: &tx.Amount,
		Script: tx.PkScript,
	}, r.Key, h.From)
}

func (r *ZookoServer) handleLookupName(data []byte, h message.Header) (message.Message, error) {
	msg := new(zMessage.LookupName)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	value, found, err := r.Names.Lookup(*msg.Lookup)
	if err != nil {
		return nil, err
	}

	return zMessage.CreateMessage(&zMessage.ResolvedName{
		Name:  msg.Lookup,
		Value: value,
		Found: &found,
	}, r.Key, h.From)
}

func (r *ZookoServer) handleRegisterName(data []byte, h message.Header) (message.Message, error) {
	msg := new(zMessage.RegisterName)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	found, err := r.Names.Register(msg)
	if err != nil {
		return nil, err
	}

	var information *string
	if !found {
		message := "Cannot register a name that is already taken."
		information = &message
	}

	return zMessage.CreateMessage(&zMessage.RegistrationResponse{
		Success:     &found,
		Information: information,
	}, r.Key, h.From)
}

func (r *ZookoServer) handleRenewName(data []byte, h message.Header) (message.Message, error) {
	msg := new(zMessage.RenewName)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	if err := r.Names.Renew(msg); err != nil {
		return nil, err
	}

	success := true
	return zMessage.CreateMessage(&zMessage.RegistrationResponse{
		Success: &success,
	}, r.Key, h.From)
}
