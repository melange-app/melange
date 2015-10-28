package identity

import (
	"fmt"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"
	"getmelange.com/dap"
	"getmelange.com/router"

	apiRouter "getmelange.com/backend/api/router"
	mIdentity "getmelange.com/backend/models/identity"

	"getmelange.com/backend/packaging"

	"airdispat.ch/identity"
	"airdispat.ch/routing"

	gdb "github.com/huntaub/go-db"
)

// Profile is a JSONObject specifying a request to create a new identity
// (or profile, I suppose).
type identityRequest struct {
	// Profile Information
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	About     string `json:"about"`
	Password  string `json:"password"`
	// AD Information
	Server  string `json:"server"`
	Tracker string `json:"tracker"`
	Alias   string `json:"alias"`
	// Identity Nickname
	Nickname string `json:"nickname"`
}

// registerServer will perform the registration for the server.
func (req *identityRequest) registerServer(server *packaging.Provider, id *identity.Identity) error {
	// Run Registration
	return (&dap.Client{
		Key:    id,
		Server: server.Key,
	}).Register(map[string][]byte{
		"name": []byte(req.FirstName + " " + req.LastName),
	})
}

// registerTracker will perform the registration with the trackers.
func (req *identityRequest) registerTracker(server *packaging.Provider, tracker *packaging.Provider, id *identity.Identity) error {
	return (&router.Router{
		Origin: id,
		TrackerList: []string{
			tracker.URL,
		},
	}).Register(id, req.Alias, map[string]routing.Redirect{
		string(routing.LookupTypeTX): routing.Redirect{
			Alias:       server.Alias,
			Fingerprint: server.Fingerprint,
			Type:        routing.LookupTypeTX,
		},
	})
}

func (req *identityRequest) registerDB(server *packaging.Provider, tracker *packaging.Provider, id *identity.Identity, env *info.Environment) (*mIdentity.Identity, error) {
	modelID, err := mIdentity.CreateIdentityFromDispatch(id, "")
	if err != nil {
		return nil, err
	}

	modelID.Nickname = req.Nickname

	// Load Server Information
	modelID.Server = server.URL
	modelID.ServerKey = server.EncryptionKey
	modelID.ServerFingerprint = server.Fingerprint
	modelID.ServerAlias = server.Alias

	_, err = env.Tables.Identity.Insert(modelID).Exec(env.Store)
	if err != nil {
		return nil, err
	}

	modelAlias := &mIdentity.Alias{
		Identity: gdb.ForeignKey(modelID),
		Location: tracker.URL,
		Username: req.Alias,
	}

	_, err = env.Tables.Alias.Insert(modelAlias).Exec(env.Store)
	if err != nil {
		return nil, err
	}

	return modelID, nil
}

// SaveIdentity will create, register, alias, and save a new Identity.
type SaveIdentity struct{}

// Post performs the specified functions.
func (i *SaveIdentity) Post(req *apiRouter.Request) framework.View {
	// Decode Body
	profileRequest := &identityRequest{}
	err := req.JSON(&profileRequest)

	if err != nil {
		fmt.Println("Error occurred while decoding body:", err)
		return framework.Error500
	}

	// Create Identity
	id, err := identity.CreateIdentity()
	if err != nil {
		fmt.Println("Error occured creating an identity:", err)
		return framework.Error500
	}

	// Get locations to register with
	server, err := req.Environment.Packager.ServerFromId(profileRequest.Server)
	if err != nil {
		fmt.Println("Error occurred getting server:", err)
		return framework.Error500
	}
	id.SetLocation(server.URL)

	tracker, err := req.Environment.Packager.TrackerFromId(profileRequest.Tracker)
	if err != nil {
		fmt.Println("Error occurred getting tracker:", err)
		return framework.Error500
	}

	// Perform registration
	err = profileRequest.registerServer(server, id)
	if err != nil {
		fmt.Println("Got error registering with the server", err)
		return framework.Error500
	}

	err = profileRequest.registerTracker(server, tracker, id)
	if err != nil {
		fmt.Println("Got error registering with the tracker", err)
		return framework.Error500
	}

	modelIdentity, err := profileRequest.registerDB(server, tracker, id, req.Environment)
	if err != nil {
		fmt.Println("Got error registering with the database", err)
		return framework.Error500
	}

	// Save as the current identity.
	err = req.Environment.Manager.SwitchIdentity(modelIdentity)
	if err != nil {
		fmt.Println("Got error switching identities", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: map[string]interface{}{
			"error": false,
		},
	}
}
