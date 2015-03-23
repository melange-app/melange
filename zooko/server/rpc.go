package main

import "github.com/melange-app/nmcd/btcjson"

type rpcServer struct {
	Username string
	Password string
	Host     string
}

func (r *rpcServer) Send(cmd btcjson.Cmd) (btcjson.Reply, error) {
	return btcjson.RpcSend(r.Username, r.Password, r.Host, cmd)
}
