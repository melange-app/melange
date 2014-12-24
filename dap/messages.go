package dap

import (
	"getmelange.com/dap/wire"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"code.google.com/p/goprotobuf/proto"
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

func CreateResponseFromBytes(data []byte, head message.Header) (*Response, error) {
	unmarsh := &wire.Response{}
	err := proto.Unmarshal(data, unmarsh)
	if err != nil {
		return nil, err
	}

	return &Response{
		Code:    unmarsh.GetCode(),
		Message: unmarsh.GetMessage(),
		Length:  unmarsh.GetLength(),
		Data:    unmarsh.GetData(),
		h:       head,
	}, nil
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
	Message *message.EncryptedMessage
	Context map[string][]byte
	h       message.Header
}

func CreateResponseMessageFromBytes(data []byte, head message.Header) (*ResponseMessage, error) {
	unmarsh := &wire.ResponseMessage{}
	err := proto.Unmarshal(data, unmarsh)
	if err != nil {
		return nil, err
	}

	enc, err := message.CreateEncryptedMessageFromBytes(unmarsh.GetData())
	if err != nil {
		return nil, err
	}

	context := make(map[string][]byte)
	for _, v := range unmarsh.GetContext() {
		context[v.GetKey()] = v.GetData()
	}

	return &ResponseMessage{
		Message: enc,
		Context: context,
		h:       head,
	}, nil
}

func CreateResponseMessage(m *message.EncryptedMessage, from *identity.Address, to *identity.Address) *ResponseMessage {
	return &ResponseMessage{
		Message: m,
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

	msgData, err := r.Message.ToBytes()
	if err != nil {
		panic("Can't marshal EncryptedMessage " + err.Error())
	}

	wire := &wire.ResponseMessage{
		Data:    msgData,
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

type RawMessage struct {
	proto.Message
	Code string
	Head message.Header
}

func (r *RawMessage) ToBytes() []byte {
	data, err := proto.Marshal(r.Message)
	if err != nil {
		panic(err.Error())
	}

	return data
}

func (r *RawMessage) Type() string           { return r.Code }
func (r *RawMessage) Header() message.Header { return r.Head }
