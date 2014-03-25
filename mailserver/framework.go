package mailserver

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"errors"
	"github.com/robfig/revel"
	"net"
	"os"
)

var ServerLocation string

func getServerLocation() string {
	s, _ := os.Hostname()
	ips, _ := net.LookupHost(s)
	return ips[0] + ":2048"
}

func Init() {
	def := getServerLocation()
	ServerLocation = revel.Config.StringDefault("server.location", def)
}

func SendAlert(r routing.Router, msgName string, from *identity.Identity, to string) error {
	addr, err := r.LookupAlias(to)
	if err != nil {
		return err
	}

	msgDescription := server.CreateMessageDescription(msgName, ServerLocation, from.Address, addr)
	err = message.SignAndSend(msgDescription, from, addr)
	if err != nil {
		return err
	}
	return nil
}

func DownloadMessage(r routing.Router, msgName string, from *identity.Identity, to string, toServer string) (*message.Mail, error) {
	addr, err := r.LookupAlias(to)
	if err != nil {
		return nil, err
	}
	addr.Location = toServer

	txMsg := server.CreateTransferMessage(msgName, from.Address, addr)
	bytes, typ, h, err := message.SendMessageAndReceive(txMsg, from, addr)
	if err != nil {
		return nil, err
	}

	if typ != wire.MailCode {
		return nil, errors.New("Wrong message type.")
	}

	return message.CreateMailFromBytes(bytes, h)
}

func DownloadPublicMail(r routing.Router, since uint64, from *identity.Identity, to string) (*server.MessageList, error) {
	addr, err := r.LookupAlias(to)
	if err != nil {
		return nil, err
	}

	txMsg := server.CreateTransferMessageList(since, from.Address, addr)
	bytes, typ, h, err := message.SendMessageAndReceive(txMsg, from, addr)
	if err != nil {
		return nil, err
	}

	if typ != wire.MessageListCode {
		return nil, errors.New("Wrong message type.")
	}

	return server.CreateMessageListFromBytes(bytes, h)
}
