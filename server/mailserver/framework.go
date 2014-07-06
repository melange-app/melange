package mailserver

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/server"
	"airdispat.ch/wire"
	"errors"
	"fmt"
	"github.com/huntaub/go-cache"
	"melange/app/models"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

var ServerLocation string = "www.airdispatch.me:2048"
var messageCache, publicCache *cache.Cache

func getServerLocation() string {
	s, _ := os.Hostname()
	ips, _ := net.LookupHost(s)
	return ips[0] + ":2048"
}

func Init() {
	// def := getServerLocation()
	// ServerLocation = revel.Config.StringDefault("server.location", def)
	// ServerLocation = "www.airdispatch.me:2048"
}

func Messages(r routing.Router,
	db models.Selectable,
	from *identity.Identity,
	fromUser *models.User,
	public bool, private bool, self bool, since int64) ([]MelangeMessage, error) {
	if messageCache == nil {
		messageCache = cache.NewCache(time.Minute * 15)
	}
	if publicCache == nil {
		publicCache = cache.NewCache(time.Minute * 5)
	}

	var out []MelangeMessage
	var err error

	if public {
		var s []*models.UserSubscription
		_, err := db.Select(&s, "select * from dispatch_subscription where userid = $1", fromUser.UserId)
		if err != nil {
			return nil, err
		}

		for _, v := range s {
			var msg *server.MessageList
			list, stale := publicCache.Get(v.Address)
			if !stale {
				msg = list.(*server.MessageList)
			} else {
				var err error
				msg, err = DownloadPublicMail(r, uint64(since), from, v.Address) // from, to)
				if err != nil {
					return nil, err
				}
				publicCache.Store(v.Address, msg)
			}
			for _, txn := range msg.Content {
				murl := fmt.Sprintf("%v::%v", txn.Location, txn.Name)
				cmsg, stale := messageCache.Get(murl)
				if !stale {
					out = append(out, CreatePluginMail(r, cmsg.(*message.Mail), from, true))
				} else {
					rmsg, err := DownloadMessage(r, txn.Name, from, v.Address, txn.Location)
					if err != nil {
						return nil, err
					}
					messageCache.Store(murl, rmsg)
					out = append(out, CreatePluginMail(r, rmsg, from, true))
				}
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
			murl := fmt.Sprintf("%v::%v", v.Location, v.Name)
			cmsg, stale := messageCache.Get(murl)
			if !stale {
				out = append(out, CreatePluginMail(r, cmsg.(*message.Mail), from, true))
			} else {
				msg, err := v.DownloadMessageFromAlert(db, r)
				if err != nil {
					return nil, err
				}
				messageCache.Store(murl, msg)
				out = append(out, CreatePluginMail(r, msg, from, false))
			}
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
			out = append(out, CreatePluginMail(r, dmsg, from, false))
		}
	}

	sort.Sort(MessageList(out))

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

func GetProfile(r routing.Router, from *identity.Identity, to string) (*message.Mail, error) {
	return DownloadMessage(r, "profile", from, to, "")
}

func DownloadMessage(r routing.Router, msgName string, from *identity.Identity, to string, toServer string) (*message.Mail, error) {
	addr, err := r.LookupAlias(to)
	if err != nil {
		return nil, err
	}
	if toServer != "" {
		addr.Location = toServer
	}

	txMsg := server.CreateTransferMessage(msgName, from.Address, addr)
	bytes, typ, h, err := message.SendMessageAndReceiveWithoutTimestamp(txMsg, from, addr)
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
