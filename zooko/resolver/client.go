package resolver

import "airdispat.ch/identity"

type Client struct {
	Origin *identity.Identity
}

func CreateClient(id *identity.Identity) *Client {
	if id == nil {
		id, _ = identity.CreateIdentity()
	}

	return &Client{
		Origin: id,
	}
}
