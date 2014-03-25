package mailserver

import (
	"airdispat.ch/message"
	"time"
)

type PluginMail struct {
	*message.Mail
}

func (p *PluginMail) Timestamp() uint64 {
	return uint64(p.Header().Timestamp)
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

func (p *PluginMail) Sender() {
}

type PluginUser struct{}
