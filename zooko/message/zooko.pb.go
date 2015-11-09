// Code generated by protoc-gen-go.
// source: source/zooko.proto
// DO NOT EDIT!

package message

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type Transaction struct {
	Id                 *string  `protobuf:"bytes,1,req,name=id" json:"id,omitempty"`
	Branch             *int32   `protobuf:"varint,2,req,name=branch" json:"branch,omitempty"`
	VerificationHashes [][]byte `protobuf:"bytes,3,rep,name=verification_hashes" json:"verification_hashes,omitempty"`
	Raw                []byte   `protobuf:"bytes,4,req,name=raw" json:"raw,omitempty"`
	BlockId            *string  `protobuf:"bytes,5,req,name=block_id" json:"block_id,omitempty"`
	Value              *int32   `protobuf:"varint,6,opt,name=value" json:"value,omitempty"`
	XXX_unrecognized   []byte   `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}

func (m *Transaction) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *Transaction) GetBranch() int32 {
	if m != nil && m.Branch != nil {
		return *m.Branch
	}
	return 0
}

func (m *Transaction) GetVerificationHashes() [][]byte {
	if m != nil {
		return m.VerificationHashes
	}
	return nil
}

func (m *Transaction) GetRaw() []byte {
	if m != nil {
		return m.Raw
	}
	return nil
}

func (m *Transaction) GetBlockId() string {
	if m != nil && m.BlockId != nil {
		return *m.BlockId
	}
	return ""
}

func (m *Transaction) GetValue() int32 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return 0
}

type ResolvedName struct {
	Found            *bool          `protobuf:"varint,1,req,name=found" json:"found,omitempty"`
	Name             *string        `protobuf:"bytes,2,req,name=name" json:"name,omitempty"`
	Value            []byte         `protobuf:"bytes,3,req,name=value" json:"value,omitempty"`
	CoinProof        []*Transaction `protobuf:"bytes,4,rep,name=coin_proof" json:"coin_proof,omitempty"`
	MessageProof     []byte         `protobuf:"bytes,5,opt,name=message_proof" json:"message_proof,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *ResolvedName) Reset()         { *m = ResolvedName{} }
func (m *ResolvedName) String() string { return proto.CompactTextString(m) }
func (*ResolvedName) ProtoMessage()    {}

func (m *ResolvedName) GetFound() bool {
	if m != nil && m.Found != nil {
		return *m.Found
	}
	return false
}

func (m *ResolvedName) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *ResolvedName) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *ResolvedName) GetCoinProof() []*Transaction {
	if m != nil {
		return m.CoinProof
	}
	return nil
}

func (m *ResolvedName) GetMessageProof() []byte {
	if m != nil {
		return m.MessageProof
	}
	return nil
}

type TransferFunds struct {
	Amount           *int64  `protobuf:"varint,1,req,name=amount" json:"amount,omitempty"`
	Id               *string `protobuf:"bytes,2,req,name=id" json:"id,omitempty"`
	Index            *uint32 `protobuf:"varint,3,req,name=index" json:"index,omitempty"`
	Script           []byte  `protobuf:"bytes,4,req,name=script" json:"script,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TransferFunds) Reset()         { *m = TransferFunds{} }
func (m *TransferFunds) String() string { return proto.CompactTextString(m) }
func (*TransferFunds) ProtoMessage()    {}

func (m *TransferFunds) GetAmount() int64 {
	if m != nil && m.Amount != nil {
		return *m.Amount
	}
	return 0
}

func (m *TransferFunds) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *TransferFunds) GetIndex() uint32 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

func (m *TransferFunds) GetScript() []byte {
	if m != nil {
		return m.Script
	}
	return nil
}

type RegistrationResponse struct {
	Success          *bool   `protobuf:"varint,1,req,name=success" json:"success,omitempty"`
	Information      *string `protobuf:"bytes,2,opt,name=information" json:"information,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RegistrationResponse) Reset()         { *m = RegistrationResponse{} }
func (m *RegistrationResponse) String() string { return proto.CompactTextString(m) }
func (*RegistrationResponse) ProtoMessage()    {}

func (m *RegistrationResponse) GetSuccess() bool {
	if m != nil && m.Success != nil {
		return *m.Success
	}
	return false
}

func (m *RegistrationResponse) GetInformation() string {
	if m != nil && m.Information != nil {
		return *m.Information
	}
	return ""
}

type RequestFunds struct {
	Address          *string `protobuf:"bytes,1,req,name=address" json:"address,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RequestFunds) Reset()         { *m = RequestFunds{} }
func (m *RequestFunds) String() string { return proto.CompactTextString(m) }
func (*RequestFunds) ProtoMessage()    {}

func (m *RequestFunds) GetAddress() string {
	if m != nil && m.Address != nil {
		return *m.Address
	}
	return ""
}

type LookupName struct {
	Lookup           *string `protobuf:"bytes,1,req,name=lookup" json:"lookup,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LookupName) Reset()         { *m = LookupName{} }
func (m *LookupName) String() string { return proto.CompactTextString(m) }
func (*LookupName) ProtoMessage()    {}

func (m *LookupName) GetLookup() string {
	if m != nil && m.Lookup != nil {
		return *m.Lookup
	}
	return ""
}

type RegisterName struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Value            []byte  `protobuf:"bytes,2,req,name=value" json:"value,omitempty"`
	NameNew          []byte  `protobuf:"bytes,3,req,name=name_new" json:"name_new,omitempty"`
	NameFirstupdate  []byte  `protobuf:"bytes,4,req,name=name_firstupdate" json:"name_firstupdate,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RegisterName) Reset()         { *m = RegisterName{} }
func (m *RegisterName) String() string { return proto.CompactTextString(m) }
func (*RegisterName) ProtoMessage()    {}

func (m *RegisterName) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *RegisterName) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *RegisterName) GetNameNew() []byte {
	if m != nil {
		return m.NameNew
	}
	return nil
}

func (m *RegisterName) GetNameFirstupdate() []byte {
	if m != nil {
		return m.NameFirstupdate
	}
	return nil
}

type RenewName struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Value            []byte  `protobuf:"bytes,2,req,name=value" json:"value,omitempty"`
	NameUpdate       []byte  `protobuf:"bytes,3,req,name=name_update" json:"name_update,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RenewName) Reset()         { *m = RenewName{} }
func (m *RenewName) String() string { return proto.CompactTextString(m) }
func (*RenewName) ProtoMessage()    {}

func (m *RenewName) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *RenewName) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *RenewName) GetNameUpdate() []byte {
	if m != nil {
		return m.NameUpdate
	}
	return nil
}

func init() {
}