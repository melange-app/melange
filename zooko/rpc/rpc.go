package rpc

import (
	"getmelange.com/zooko/account"
	"github.com/melange-app/nmcd/btcjson"
)

// Server represents
type Server struct {
	Username string
	Password string
	Host     string
}

const (
	// This allows for four "transaction-units to occur":
	// - name_new + fee
	// - name_firstupdate + fee
	endowmentAmount = 5e5 * 4
)

// Endow will transfer money for a single Namecoin name registration
// (for 12 months) to a specified address. It will return the
// transaction details so that light clients can keep track of their
// UTXO.
func (r *Server) Endow(address string) (*account.UTXO, error) {
	cmd, err := btcjson.NewSendToAddressCmd(nil, address, endowmentAmount, nil)
	if err != nil {
		return err
	}

	reply, err := r.Send(cmd)
	if err != nil {
		return err
	}

	txid := reply.(string)

	return &account.UTXO{
		TXID:   txid,
		Output: 1,
		Amount: endowmentAmount,
	}, nil
}

// Broadcast will broadcast a raw hex-encoded transaction to the
// Namecoin network.
func (r *Server) Broadcast(tx string) error {

}

// Send will send a raw btcjson.Cmd to the server.
func (r *Server) Send(cmd btcjson.Cmd) (btcjson.Reply, error) {
	return btcjson.RpcSend(r.Username, r.Password, r.Host, cmd)
}
