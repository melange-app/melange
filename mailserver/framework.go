package mailserver

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"errors"
	"github.com/airdispatch/dpl"
	"github.com/robfig/revel"
	"melange/app/models"
	"net"
	"os"
	"strings"
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

func Messages(r routing.Router,
	db models.Selectable,
	from *identity.Identity,
	fromUser *models.User,
	public bool, private bool, self bool, since int64) ([]dpl.Message, error) {

	var out []dpl.Message
	var err error

	if public {
		var s []*models.UserSubscription
		_, err := db.Select(&s, "select * from dispatch_subscription where userid = $1", fromUser.UserId)
		if err != nil {
			return nil, err
		}

		for _, v := range s {
			msg, err := DownloadPublicMail(r, uint64(since), from, v.Address) // from, to)
			if err != nil {
				return nil, err
			}
			for _, txn := range msg.Content {
				rmsg, err := DownloadMessage(r, txn.Name, nil, v.Address, txn.Location)
				if err != nil {
					return nil, err
				}
				out = append(out, &PluginMail{rmsg})
			}
		}
	}

	if private {
		var ale []*models.Alert
		_, err = db.Select(&ale, "select * from dispatch_alerts where \"to\" = $1 and timestamp > $2", from.Address.String(), since)
		if err != nil {
			return nil, err
		}
		for _, v := range ale {
			msg, err := v.DownloadMessageFromAlert(db, r)
			if err != nil {
				return nil, err
			}
			out = append(out, &PluginMail{msg})
		}
	}

	if self {
		var msg []*models.Message
		_, err = db.Select(&msg, "select * from dispatch_messages where \"from\" = $1 and timestamp > $2", from.Address.String(), since)
		if err != nil {
			return nil, err
		}
		for _, v := range msg {
			dmsg, err := v.ToDispatch(db, strings.Split(v.To, ",")[0])
			if err != nil {
				return nil, err
			}
			out = append(out, &PluginMail{dmsg})
		}
	}

	return out, nil
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

func DownloadMessageList(r routing.Router, m *server.MessageList, from *identity.Identity, to string) ([]*message.Mail, error) {
	output := make([]*message.Mail, len(m.Content))
	for i, v := range m.Content {
		var err error
		output[i], err = DownloadMessage(r, v.Name, from, to, v.Location)
		if err != nil {
			return nil, err
		}
	}
	return output, nil
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

	if len(bytes) == 0 {
		// No messages available.
		return &server.MessageList{
			Content: make([]*server.MessageDescription, 0),
		}, nil
	}
	return server.CreateMessageListFromBytes(bytes, h)
}
