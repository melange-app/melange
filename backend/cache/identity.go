package cache

import (
	"getmelange.com/backend/connect"
	"getmelange.com/backend/models/identity"
	"getmelange.com/router"
)

func (m *Manager) currentIdentity() (*identity.Identity, error) {
	data, err := m.Store.GetDefault("current_identity", "")
	if err != nil {
		return nil, err
	}

	if data == "" {
		return nil, nil
	}

	result := &identity.Identity{}
	m.Tables.Identity.Get().Where("fingerprint", data).One(m.Store, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *Manager) setCurrentIdentity(newIdentity *identity.Identity) error {
	byFingerprint := []byte(newIdentity.Fingerprint)
	return m.Store.SetBytes("current_identity", byFingerprint)
}

func (m *Manager) loadIdentity(id *identity.Identity) error {
	dapClient, err := id.Client(m.Store)
	if err != nil {
		return err
	}

	router := &router.Router{
		Origin: dapClient.Key,
		TrackerList: []string{
			"localhost:2048",
		},
	}

	m.Identity = id
	m.Client = &connect.Client{
		Client: dapClient,
		Router: router,
		Origin: dapClient.Key,
	}
	m.Fetcher = CreateFetcher(m.Client, m.Cache, m.Tables, m.Store)
	m.HasIdentity = true

	return nil
}

// SwitchIdentity will change the current identity being used in the
// cache, fetcher, and client.
func (m *Manager) SwitchIdentity(newIdentity *identity.Identity) error {
	if m.Fetcher != nil {
		m.Fetcher.Stop()
	}

	m.Cache.Clear()

	if err := m.setCurrentIdentity(newIdentity); err != nil {
		m.HasIdentity = false
		m.Identity = nil
		return err
	}

	// Stopping here is a bad error because we have no identity
	// perform requests with.

	return m.loadIdentity(newIdentity)
}
