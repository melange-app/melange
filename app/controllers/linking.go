package controllers

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/routing"
	"getmelange.com/dap"

	"airdispat.ch/crypto"
	"code.google.com/p/go-uuid/uuid"
	"getmelange.com/app/models"
	"getmelange.com/router"

	gdb "github.com/huntaub/go-db"
)

// Requestor
// 1. Send Link Request
// 2. Await Acceptance
// 3. Download Identity

func (r *RealtimeHandler) RequestLink(d interface{}) (string, map[string]interface{}) {
	// Cast to map
	data, ok := d.(map[string]interface{})
	if !ok {
		return "requestedLink", map[string]interface{}{
			"error": "malformed request, requires map",
		}
	}

	// Create Temporary Identity for Router
	tempId, err := identity.CreateIdentity()
	if err != nil {
		fmt.Println("Error creating identity", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	addrObj, ok := data["address"]
	if !ok {
		return "requestedLink", map[string]interface{}{
			"error": "malformed request, need address",
		}
	}

	addr, ok := addrObj.(string)
	if !ok {
		return "requestedLink", map[string]interface{}{
			"error": "malformed request, address must be string",
		}
	}

	// Create Router
	router := &router.Router{
		Origin: tempId,
	}

	// Lookup Server Location
	srv, err := router.LookupAlias(addr, routing.LookupTypeTX)
	if err != nil {
		fmt.Println("Error looking up server", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Lookup User Location
	person, err := router.LookupAlias(addr, routing.LookupTypeMAIL)
	if err != nil {
		fmt.Println("Error looking up person", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Create the Link Client
	lc, err := dap.CreateLinkClient(person, srv)
	if err != nil {
		fmt.Println("Error creating Link Client", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Create the Verification Code
	vc, err := lc.GetVerificationCode()
	if err != nil {
		fmt.Println("Error getting VC", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Display the Verification Code
	r.dataChan <- map[string]interface{}{
		"type": "linkVerification",
		"data": map[string]interface{}{
			"code": string(vc),
		},
	}

	// Send the Link Request
	err = lc.LinkRequest(vc)
	if err != nil {
		// This is probably related to the other side not being enabled in time...
		fmt.Println("Error sending link request", err)
		return "requestedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		stop := time.NewTimer(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				stop.Stop()

				recvId, err := lc.LinkGetIdentity()
				if err == nil {
					// Great success!
					fmt.Println("Got id", recvId.Address.String(), "for addr", addr)

					modelId, err := models.CreateIdentityFromDispatch(recvId, "")
					if err != nil {
						fmt.Println("Got error creating ID", err)
						r.dataChan <- map[string]interface{}{
							"type": "linkedIdentity",
							"data": map[string]interface{}{
								"error": err.Error(),
							},
						}
					}

					// Load the Server Information
					modelId.Server = srv.Location
					modelId.ServerFingerprint = srv.String()
					modelId.ServerKey = hex.EncodeToString(
						crypto.RSAToBytes(
							srv.EncryptionKey,
						),
					)

					modelId.Nickname = fmt.Sprintf(
						"Imported %s",
						time.Now().Format("3:04PM Jan 2, 2006"),
					)

					// TODO: srv.Alias is likely to be the empty string
					if srv.Alias == "" {
						// Let's get the server alias!
						fmt.Println("The linked server has no alias.")
					}

					modelId.ServerAlias = srv.Alias

					_, err = r.Tables["identity"].Insert(modelId).Exec(r.Store)
					if err != nil {
						fmt.Println("Got error saving ID", err)
						r.dataChan <- map[string]interface{}{
							"type": "linkedIdentity",
							"data": map[string]interface{}{
								"error": err.Error(),
							},
						}
					}

					addrComponents := strings.Split(addr, "@")
					location := ""
					username := ""
					if len(addrComponents) == 2 {
						location = addrComponents[1]
						username = addrComponents[0]
					} else {
						fmt.Println(addr, "isn't in the correct form for linking.")
					}

					_, err = r.Tables["alias"].Insert(
						&models.Alias{
							Identity: gdb.ForeignKey(modelId),
							Location: location,
							Username: username,
						},
					).Exec(r.Store)
					if err != nil {
						fmt.Println("Got error saving alias", err)
						r.dataChan <- map[string]interface{}{
							"type": "linkedIdentity",
							"data": map[string]interface{}{
								"error": err.Error(),
							},
						}
					}

					r.dataChan <- map[string]interface{}{
						"type": "linkedIdentity",
						"data": map[string]interface{}{
							"fingerprint": modelId.Fingerprint,
						},
					}

					return
				}
			case <-stop.C:
				fmt.Println("Timed out of Receive Identity")
				ticker.Stop()

				r.dataChan <- map[string]interface{}{
					"type": "linkedIdentity",
					"data": map[string]interface{}{
						"error": "Timed out waiting to get ID.",
					},
				}
				return
			}
		}
	}()

	return "requestedLink", map[string]interface{}{}
}

// Requestee
// 1. Enable Account for Linking
// 2. Await Link Request
// 3. Accept Link Request

func (r *RealtimeHandler) getIdentityWithFingerprint(fp string) (*models.Identity, error) {
	// Grab the Identity from the Table
	obj := &models.Identity{}
	err := r.Tables["identity"].Get().Where("fingerprint", fp).One(r.Store, obj)

	return obj, err
}

type linkRequest struct {
	*dap.PendingLinkRequest
	Client *dap.Client
}

func (r *RealtimeHandler) StartLink(d interface{}) (string, map[string]interface{}) {
	// Cast to map[string]interface{}
	data, ok := d.(map[string]interface{})
	if !ok {
		return "startedLink", map[string]interface{}{
			"error": "malformed request, requires map",
		}
	}

	// Load the Correct Address
	id, ok := data["fingerprint"]
	if !ok {
		return "startedLink", map[string]interface{}{
			"error": "malformed request, requires fingerprint",
		}
	}

	idStr, ok := id.(string)
	if !ok {
		return "startedLink", map[string]interface{}{
			"error": "malformed request, fingerprint must be string",
		}
	}

	modelId, err := r.getIdentityWithFingerprint(idStr)
	if err != nil {
		fmt.Println("Got error getting identity", err)
		return "startedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Create the DAP Client
	client, err := DAPClientFromID(modelId, r.Store)
	if err != nil {
		fmt.Println("Got error getting DAP Client", err)
		return "startedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	err = client.EnableLink()
	if err != nil {
		fmt.Println("Got error enabling link", err)
		return "startedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	go func() {
		request, err := client.WaitForLinkRequest()
		if err != nil {
			fmt.Println("Got error waiting for request", err)
			r.dataChan <- map[string]interface{}{
				"type": "startedLink",
				"data": map[string]interface{}{
					"error": err.Error(),
				},
			}
		}

		id := uuid.New()

		r.dataChan <- map[string]interface{}{
			"type": "startedLink",
			"data": map[string]interface{}{
				"code": string(request.Verification),
				"uuid": id,
			},
		}

		r.requestsLock.Lock()
		r.requests[id] = &linkRequest{
			PendingLinkRequest: request,
			Client:             client,
		}
		r.requestsLock.Unlock()
	}()

	return "waitingForLinkRequest", map[string]interface{}{}
}

func (r *RealtimeHandler) AcceptLink(d interface{}) (string, map[string]interface{}) {
	r.requestsLock.RLock()
	defer r.requestsLock.RUnlock()

	data, ok := d.(map[string]interface{})
	if !ok {
		return "acceptedLink", map[string]interface{}{
			"error": "malformed request, must send map",
		}
	}

	id, ok := data["uuid"]
	if !ok {
		return "acceptedLink", map[string]interface{}{
			"error": "malformed request, requires uuid",
		}
	}

	idStr, ok := id.(string)
	if !ok {
		return "acceptedLink", map[string]interface{}{
			"error": "malformed request, uuid must be string",
		}
	}

	request, ok := r.requests[idStr]
	if !ok {
		return "acceptedLink", map[string]interface{}{
			"error": "malformed request, that id does not exist",
		}
	}

	err := request.Client.LinkAcceptRequest(request.PendingLinkRequest)
	if err != nil {
		fmt.Println("Error accepting request", err)
		return "acceptedLink", map[string]interface{}{
			"error": err.Error(),
		}
	}

	return "acceptedLink", map[string]interface{}{}
}
