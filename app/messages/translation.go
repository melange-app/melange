package messages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gdb "github.com/huntaub/go-db"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"getmelange.com/app/models"
	"getmelange.com/dap"
)

func translateComponents(comp message.ComponentList) map[string]*models.JSONComponent {
	out := make(map[string]*models.JSONComponent)

	for _, v := range comp {
		out[v.Name] = &models.JSONComponent{
			String: string(v.Data),
		}
	}

	return out
}

func translateMessageWithContext(
	r routing.Router,
	from *identity.Identity,
	public bool,
	context map[string][]byte,
	msg *message.Mail,
) []*models.JSONMessage {
	obj := translateMessage(r, from, public, msg)

	out := make(map[string]string)
	for key, v := range context {
		out[key] = string(v)
	}

	obj[0].Context = out

	return obj
}

func translateModel(
	r routing.Router,
	cmpTable gdb.Table,
	store gdb.Executor,
	v *models.Message, d *dap.Client,
	myProfile *models.Profile,
	myAlias *models.Alias,
) (*models.JSONMessage, error) {
	comps := make([]*models.Component, 0)
	realErr := cmpTable.Get().Where("message", v.Id).All(store, &comps)
	if realErr != nil {
		return nil, realErr
	}

	mlgComps := make(map[string]*models.JSONComponent)
	for _, c := range comps {
		mlgComps[c.Name] = &models.JSONComponent{
			Binary: c.Data,
			String: string(c.Data),
		}
	}

	// Download Profile
	toAddrs := strings.Split(v.To, ",")

	profiles := make([]*models.JSONProfile, 0)
	for _, j := range toAddrs {
		if j == "" {
			continue
		}

		_, addr, err := GetAddresses(r, &models.Address{
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

	return &models.JSONMessage{
		Name: v.Name,
		Date: time.Unix(v.Date, 0),
		// To and From Info
		From: &models.JSONProfile{
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

func translateMessage(
	r routing.Router,
	from *identity.Identity,
	public bool,
	msg ...*message.Mail,
) []*models.JSONMessage {
	out := make([]*models.JSONMessage, len(msg))

	for i, v := range msg {
		var profile *models.JSONProfile

		var err error

		profile, err = translateProfile(r, from, v.Header().From.String(), v.Header().Alias)
		if err != nil {
			fmt.Println("Couldn't get profile", err)

			name := v.Header().From.String()
			if v.Header().Alias != "" {
				name = v.Header().Alias
			}

			profile = &models.JSONProfile{
				Name:        name,
				Fingerprint: v.Header().From.String(),
				Avatar:      defaultProfileImage(v.Header().From), // haha
				Alias:       v.Header().Alias,
			}
		}

		named := v.Name
		if named == "" {
			named = fmt.Sprintf("__%d", time.Now().Unix())
		}

		out[i] = &models.JSONMessage{
			Name:       named,
			Date:       time.Unix(v.Header().Timestamp, 0),
			From:       profile,
			Public:     public,
			Components: translateComponents(v.Components),
			Context:    nil,
		}
	}

	return out
}

func translateProfile(
	r routing.Router,
	from *identity.Identity,
	fp string,
	alias string,
) (*models.JSONProfile, error) {
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

	refresh := func() (*models.JSONProfile, error) {
		profile, err := getProfile(r, from, fp, alias)
		if err != nil {
			switch t := err.(type) {
			case (*adErrors.Error):
				if t.Code == 5 {
					// Profile doesn't exist.
					// Return default profile.
					return &models.JSONProfile{
						Name:        alias,
						Avatar:      noImage,
						Alias:       alias,
						Fingerprint: fp,
					}, nil
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

		// This is an AirDispatch image.
		if strings.Contains(avatar, "@") {
			avatar = fmt.Sprintf("http://data.melange:7776/%s", avatar)
		}

		p := &models.JSONProfile{
			Name:        name,
			Avatar:      avatar,
			Alias:       alias,
			Fingerprint: fp,
		}

		return p, nil
	}

	profile, err := refresh()
	if err != nil {
		return nil, err
	}

	majorStore.AddProfile(profile, refresh)

	return profile, nil
}

func defaultProfileImage(from *identity.Address) string {
	return fmt.Sprintf("http://robohash.org/%s.png?bgset=bg2", from.String())
}

func LoadContactProfile(
	r routing.Router,
	c *models.Contact,
	from *identity.Identity,
) error {
	currentAddress := c.Identities[0]

	fp := currentAddress.Fingerprint
	if fp == "" {
		temp, err := r.LookupAlias(currentAddress.Alias, routing.LookupTypeMAIL)
		if err != nil {
			return err
		}
		fp = temp.String()

	}

	profile, err := translateProfile(r, from, fp, currentAddress.Alias)
	if err != nil {
		return err
	}

	c.Profile = profile
	return nil
}
