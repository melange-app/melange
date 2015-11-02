package message

import (
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
)

const (
	LookupNameCode   = "NMC-LKP"
	ResolvedNameCode = "NMC-RSD"
)

type LookupNameMessage struct {
	Name string
	H    message.Header
}

func CreateLookupNameMessageFromBytes(by []byte, h message.Header) *LookupNameMessage {
	return &LookupNameMessage{
		Name: string(by),
		H:    h,
	}
}

func (l *LookupNameMessage) Type() string           { return LookupNameCode }
func (l *LookupNameMessage) Header() message.Header { return l.H }

func (l *LookupNameMessage) ToBytes() []byte {
	return []byte(l.Name)
}

type NamecoinTransaction struct {
	TxId               string
	Branch             int32
	VerificationHashes [][]byte
	Raw                []byte
	BlockId            string
	Value              int32
}

type ResolvedNameMessage struct {
	Transactions []NamecoinTransaction
	Found        bool
	H            message.Header
}

func CreateResolvedNameMessageFromBytes(by []byte, h message.Header) (*ResolvedNameMessage, error) {
	newMsg := new(ResolvedName)
	err := proto.Unmarshal(by, newMsg)
	if err != nil {
		return nil, err
	}

	var intx []NamecoinTransaction
	for _, v := range newMsg.Transactions {
		value := int32(0)
		if v.Value != nil {
			value = *v.Value
		}

		intx = append(intx, NamecoinTransaction{
			TxId:               *v.Id,
			Branch:             *v.Branch,
			VerificationHashes: v.VerificationHashes,
			Raw:                v.Raw,
			BlockId:            *v.BlockId,
			Value:              value,
		})
	}

	return &ResolvedNameMessage{
		Found:        *newMsg.Found,
		Transactions: intx,
		H:            h,
	}, nil
}

func (r *ResolvedNameMessage) Type() string           { return ResolvedNameCode }
func (r *ResolvedNameMessage) Header() message.Header { return r.H }

func (r *ResolvedNameMessage) ToBytes() []byte {
	var outtx []*Transaction
	for _, v := range r.Transactions {
		outtx = append(outtx, &Transaction{
			Id:                 &v.TxId,
			Branch:             &v.Branch,
			VerificationHashes: v.VerificationHashes,
			Raw:                v.Raw,
			BlockId:            &v.BlockId,
			Value:              &v.Value,
		})
	}
	msg := &ResolvedName{
		Transactions: outtx,
		Found:        &r.Found,
	}

	by, err := proto.Marshal(msg)
	if err != nil {
		panic("Unable to Marshal Resolved Name Message:" + err.Error())
	}

	return by
}
