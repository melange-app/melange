package models

import (
	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"bytes"
	"github.com/coopernurse/gorp"
)

type Identity struct {
	IdentityId  int
	UserId      int
	Fingerprint string
	Data        []byte
}

func IdentityFromFingerprint(fgp string, dbMap *gorp.DbMap) *identity.Identity {
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
