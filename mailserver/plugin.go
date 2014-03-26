package mailserver

import (
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"github.com/airdispatch/dpl"
	"net/url"
	"time"
)

type PluginMail struct {
	*message.Mail
}

func (p *PluginMail) Get(field string) ([]byte, error) {
	return p.Components.GetComponent(field), nil
}

func (p *PluginMail) Has(field string) bool {
	return p.Components.HasComponent(field)
}

func (p *PluginMail) Created() time.Time {
	return time.Unix(p.Header().Timestamp, 0)
}

func (p *PluginMail) Sender() dpl.User {
	return &PluginUser{
		loaded: p.Header().From,
	}
}

// Abstract Getting User's Profile
type PluginUser struct {
	loaded *identity.Address
}

func (p *PluginUser) Name() string {
	return "Name TODO"
}

func (p *PluginUser) DisplayAddress() string {
	return "Display Address TODO"
}

func (p *PluginUser) Address() string {
	return "Address TODO"
}

func (p *PluginUser) Avatar() *url.URL {
	u, _ := url.Parse("http://google.com")
	return u
}

func (p *PluginUser) Profile() *url.URL {
	u, _ := url.Parse("http://google.com")
	return u
}
