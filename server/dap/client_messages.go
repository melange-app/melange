package dap

import (
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
)

type RawMessage struct {
	proto.Message
	Code string
	Head message.Header
}

func (r *RawMessage) ToBytes() []byte {
	data, err := proto.Marshal(r)
	if err != nil {
		panic(err.Error())
	}

	return data
}

func (r *RawMessage) Type() string           { return r.Code }
func (r *RawMessage) Header() message.Header { return r.Head }
