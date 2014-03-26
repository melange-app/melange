package models

import (
	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"bytes"
)

type Identity struct {
	IdentityId  int
	UserId      int
	Fingerprint string
	Data        []byte
}

func NewIdentityForUser(u *User) (*Identity, error) {
	id, err := identity.CreateIdentity()
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	_, err = id.GobEncodeKey(&b)
	if err != nil {
		return nil, err
	}

	return &Identity{
		UserId:      u.UserId,
		Fingerprint: id.Address.String(),
		Data:        b.Bytes(),
	}, nil
}

func UserIdentities(u *User, dbMap Selectable) ([]*identity.Identity, error) {
	var id []*Identity
	_, err := dbMap.Select(&id, "select * from dispatch_identity where userid = $1", u.UserId)
	if err != nil {
		return nil, err
	}
	var out []*identity.Identity = make([]*identity.Identity, len(id))
	for i, v := range id {
		did, err := v.ToDispatch()
		if err != nil {
			continue
		}
		out[i] = did
	}
	return out, nil
}

func IdentityFromFingerprint(fgp string, dbMap Selectable) *identity.Identity {
	var i []*Identity
	_, err := dbMap.Select(&i, "select * from dispatch_identity where fingerprint = $1", fgp)
	if err != nil {
		// Handle Error?
		return nil
	}

	if len(i) != 1 {
		return nil
	}

	outId, err := i[0].ToDispatch()
	if err != nil {
		// Handle Error?
		return nil
	}

	return outId
}

func (i *Identity) ToDispatch() (*identity.Identity, error) {
	buf := bytes.NewBuffer(i.Data)
	return identity.GobDecodeKey(buf)
}

func (a *Identity) Register(r routing.Router) error {
	adId, err := a.ToDispatch()
	if err != nil {
		return err
	}

	return r.Register(adId)
}
