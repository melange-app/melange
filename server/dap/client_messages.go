package dap

import (
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
)

type ClientMessage struct {
	proto.Message
	Code string
	Head message.Header
}

func (r *ClientMessage) ToBytes() []byte {
	data, err := proto.Marshal(r)
	if err != nil {
		panic(err.Error())
	}

	return data
}

func (r *ClientMessage) Type() string           { return r.Code }
func (r *ClientMessage) Header() message.Header { return r.Head }
