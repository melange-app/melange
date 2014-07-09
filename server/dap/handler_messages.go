package dap

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
	"melange/dap/wire"
)

func CreateResponse(code int, msg string, from *identity.Address, to *identity.Address, data ...message.Message) []message.Message {
	return append([]message.Message{
		&Response{
			Code:    uint32(code),
			Message: msg,
			Length:  uint64(len(data)),
			h:       message.CreateHeader(from, to),
		},
	}, data...)
}

func CreateDataResponse(code int, msg string, from *identity.Address, to *identity.Address, data []byte) []message.Message {
	return []message.Message{
		&Response{
			Code:    uint32(code),
			Message: msg,
			Data:    data,
			h:       message.CreateHeader(from, to),
		},
	}
}

type Response struct {
	Code    uint32
	Message string
	Length  uint64
	Data    []byte
	h       message.Header
}

func (r *Response) Header() message.Header {
	return r.h
}

func (r *Response) ToBytes() []byte {
	wire := &wire.Response{
		Code:    &r.Code,
		Message: &r.Message,
		Length:  &r.Length,
		Data:    r.Data,
	}
	data, err := proto.Marshal(wire)
	if err != nil {
		panic("Can't marshal response " + err.Error())
	}
	return data
}

func (r *Response) Type() string {
	return wire.ResponseCode
}

type ResponseMessage struct {
	Message []byte
	Context map[string][]byte
	h       message.Header
}

func CreateResponseMessage(m *message.EncryptedMessage, from *identity.Address, to *identity.Address) *ResponseMessage {
	data, err := m.ToBytes()
	if err != nil {
		panic("Can't marshal EncryptedMessage " + err.Error())
	}

	return &ResponseMessage{
		Message: data,
		Context: make(map[string][]byte),
		h:       message.CreateHeader(from, to),
	}
}

func (r *ResponseMessage) Header() message.Header {
	return r.h
}

func (r *ResponseMessage) ToBytes() []byte {
	context := make([]*wire.Data, 0)
	for key, value := range r.Context {
		newKey := key
		context = append(context, &wire.Data{
			Key:  &newKey,
			Data: value,
		})
	}

	wire := &wire.ResponseMessage{
		Data:    r.Message,
		Context: context,
	}
	data, err := proto.Marshal(wire)
	if err != nil {
		panic("Can't marshal response " + err.Error())
	}
	return data
}

func (r *ResponseMessage) Type() string {
	return wire.ResponseMessageCode
}
