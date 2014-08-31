package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"

	"getmelange.com/dap"

	"getmelange.com/router"
	gdb "github.com/huntaub/go-db"
)

// JSON Encoding
type JSONMessageList []*JSONMessage

func (m JSONMessageList) Len() int               { return len(m) }
func (m JSONMessageList) Less(i int, j int) bool { return m[i].Date.After(m[j].Date) }
func (m JSONMessageList) Swap(i int, j int)      { m[i], m[j] = m[j], m[i] }

type JSONMessage struct {
	Name       string                    `json:"name"`
	Date       time.Time                 `json:"date"`
	From       *JSONProfile              `json:"from"`
	To         []*JSONProfile            `json:"to"`
	Public     bool                      `json:"public"`
	Self       bool                      `json:"self"`
	Components map[string]*JSONComponent `json:"components"`
	Context    map[string]string         `json:"context"`
}

type JSONComponent struct {
	Binary []byte `json:"binary"`
	String string `json:"string"`
}

type JSONProfile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Alias       string `json:"alias"`
	Fingerprint string `json:"fingerprint"`
}

func (m *JSONMessage) ToModel(from *identity.Identity) (*Message, []*Component) {
	toAddrs := ""
	for i, v := range m.To {
		if i != 0 {
			toAddrs += ","
		}
		toAddrs += v.Alias
	}

	message := &Message{
		Name: m.Name,
		// Address
		To:   toAddrs,
		From: from.Address.String(),
		// Meta
		Date:     m.Date.Unix(),
		Incoming: false,
		Alert:    !m.Public,
	}

	components := make([]*Component, len(m.Components))
	i := 0
	for key, v := range m.Components {
		components[i] = &Component{
			Name: key,
		}

		if len(v.Binary) == 0 {
			components[i].Data = []byte(v.String)
		} else {
			components[i].Data = v.Binary
		}

		i++
	}

	return message, components
}

func (m *JSONMessage) ToDispatch(from *identity.Identity) (*message.Mail, []*identity.Address, error) {
	r := router.Router{
		Origin: from,
	}

	addrs := make([]*identity.Address, len(m.To))
	for i, v := range m.To {
		var err error
		addrs[i], err = r.LookupAlias(v.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return nil, nil, err
		}
	}

	mail := message.CreateMail(from.Address, m.Date, addrs...)

	for key, v := range m.Components {
		mail.Components.AddComponent(message.Component{
			Name: key,
			Data: []byte(v.String),
		})
	}

	return mail, addrs, nil
}

func translateComponents(comp message.ComponentList) map[string]*JSONComponent {
	out := make(map[string]*JSONComponent)

	for _, v := range comp {
		out[v.Name] = &JSONComponent{
			String: string(v.Data),
		}
	}

	return out
}

func translateMessageWithContext(r routing.Router, from *identity.Identity, public bool, context map[string][]byte, msg *message.Mail) []*JSONMessage {
	obj := translateMessage(r, from, public, msg)

	out := make(map[string]string)
	for key, v := range context {
		out[key] = string(v)
	}

	obj[0].Context = out

	return obj
}

func translateModel(r routing.Router, cmpTable gdb.Table, store gdb.Executor, v *Message, d *dap.Client, myProfile *Profile, myAlias *Alias) (*JSONMessage, error) {
	now := time.Now()
	comps := make([]*Component, 0)
	realErr := cmpTable.Get().Where("message", v.Id).All(store, &comps)
	if realErr != nil {
		return nil, realErr
	}
	now = time.Now()
	fmt.Println("Getting Components takes", time.Now().Sub(now))
	defer func() {
		fmt.Println("Getting Everything else takes", time.Now().Sub(now))
	}()

	mlgComps := make(map[string]*JSONComponent)
	for _, c := range comps {
		mlgComps[c.Name] = &JSONComponent{
			Binary: c.Data,
			String: string(c.Data),
		}
	}

	// Download Profile
	toAddrs := strings.Split(v.To, ",")

	profiles := make([]*JSONProfile, 0)
	for _, j := range toAddrs {
		if j == "" {
			continue
		}

		_, addr, err := getAddresses(r, &Address{
			Alias: j,
		})
		if err != nil {
			fmt.Println("Couldn't get fp for", j, err)
			continue
		}

		p, err := translateProfile(r, d.Key, addr.String(), j)
		if err != nil {
			fmt.Println("Couldn't get profile for", j, err)
		}

		profiles = append(profiles, p)
	}

	return &JSONMessage{
		Name: v.Name,
		Date: time.Unix(v.Date, 0),
		// To and From Info
		From: &JSONProfile{
			Name:        myProfile.Name,
			Avatar:      myProfile.Image,
			Alias:       myAlias.String(),
			Fingerprint: d.Key.Address.String(),
		},
		To: profiles,
		// Components
		Components: mlgComps,
		// Meta
		Self:   true,
		Public: !v.Alert,
	}, nil
}

func translateMessage(r routing.Router, from *identity.Identity, public bool, msg ...*message.Mail) []*JSONMessage {
	out := make([]*JSONMessage, len(msg))

	for i, v := range msg {
		var profile *JSONProfile

		var err error

		profile, err = translateProfile(r, from, v.Header().From.String(), v.Header().Alias)
		if err != nil {
			fmt.Println("Couldn't get profile", err)

			name := v.Header().From.String()
			if v.Header().Alias != "" {
				name = v.Header().Alias
			}

			profile = &JSONProfile{
				Name:        name,
				Fingerprint: v.Header().From.String(),
				Avatar:      defaultProfileImage(v.Header().From), // haha
				Alias:       v.Header().Alias,
			}
		}

		out[i] = &JSONMessage{
			Name:       "",
			Date:       time.Unix(v.Header().Timestamp, 0),
			From:       profile,
			Public:     public,
			Components: translateComponents(v.Components),
			Context:    nil,
		}
	}

	return out
}

func translateProfile(r routing.Router, from *identity.Identity, fp string, alias string) (*JSONProfile, error) {
	now := time.Now()
	defer func() {
		fmt.Println("Got profile", time.Now().Sub(now))
	}()

	if profile, ok := majorStore.RetrieveProfile(fp); ok {
		return profile, nil
	}

	if alias == "" {
		return nil, errors.New("Can't get profile without alias support.")
	}

	if fp == "" {
		fmt.Println("Image won't be correct for no fp lookup with alias", alias)
	}
	noImage := defaultProfileImage(identity.CreateAddressFromString(fp))

	refresh := func() (*JSONProfile, error) {
		return translateProfile(r, from, fp, alias)
	}

	profile, err := getProfile(r, from, fp, alias)
	if err != nil {
		switch t := err.(type) {
		case (*adErrors.Error):
			if t.Code == 5 {
				p := &JSONProfile{
					Name:        alias,
					Avatar:      noImage,
					Alias:       alias,
					Fingerprint: fp,
				}
				majorStore.AddProfile(p, refresh)
				return p, nil
			}
			return nil, err
		case error:
			return nil, err
		}
	}

	name := profile.Components.GetStringComponent("airdispat.ch/profile/name")
	if name == "" {
		name = alias
	}

	avatar := profile.Components.GetStringComponent("airdispat.ch/profile/avatar")
	if avatar == "" {
		avatar = noImage
	}

	p := &JSONProfile{
		Name:        name,
		Avatar:      avatar,
		Alias:       alias,
		Fingerprint: fp,
	}

	majorStore.AddProfile(p, refresh)

	return p, nil
}

func defaultProfileImage(from *identity.Address) string {
	return fmt.Sprintf("http://robohash.org/%s.png?bgset=bg2", from.String())
}
