package message

import "code.google.com/p/goprotobuf/proto"

const (
	// Server message types.
	TypeTransaction          = "ZK-TX"
	TypeResolvedName         = "ZK-RS"
	TypeListNames            = "ZK-LS"
	TypeTransferFunds        = "ZK-TF"
	TypeRegistrationResponse = "ZK-RR"

	// Client message types.
	TypeRequestFunds = "ZK-RF"
	TypeLookupName   = "ZK-LN"
	TypeRegisterName = "ZK-RN"
	TypeRenewName    = "ZK-RW"
)

func getMessageType(m proto.Message) string {
	switch m.(type) {
	// Server messages
	case *Transaction:
		return TypeTransaction
	case *ResolvedName:
		return TypeResolvedName
	case *ListName:
		return TypeListNames
	case *TransferFunds:
		return TypeTransferFunds
	case *RegistrationResponse:
		return TypeRegistrationResponse

		// Client messsage
	case *RequestFunds:
		return TypeRequestFunds
	case *LookupName:
		return TypeLookupName
	case *RegisterName:
		return TypeRegisterName
	case *RenewName:
		return TypeRenewName

	default:
		panic("zooko: cannot use getMessageType with a non-zooko message")
	}
}
