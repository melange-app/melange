package rpc

import (
	"errors"

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
	// This allows for three "transaction-units to occur":
	// - name_new + fee
	// - name_firstupdate + fee
	endowmentAmount = account.NameMinimumBalance * 3
)

var (
	errNilReply = errors.New("zooko/rpc: nil response from namecoin daemon")
)

// Send will send a raw btcjson.Cmd to the server.
func (r *Server) Send(cmd btcjson.Cmd) (btcjson.Reply, error) {
	return btcjson.RpcSend(r.Username, r.Password, r.Host, cmd)
}
