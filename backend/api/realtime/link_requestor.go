package realtime

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"airdispat.ch/identity"
	"getmelange.com/dap"

	"airdispat.ch/crypto"

	"getmelange.com/backend/connect"
	mIdentity "getmelange.com/backend/models/identity"
	"getmelange.com/router"

	gdb "github.com/huntaub/go-db"
)

const (
	requestedLink        = "requestedLink"
	linkVerificationType = "linkVerification"
	linkDoneType         = "linkedIdentity"
)

type requestLinkResponse struct {
	Code string `json:"code"`
}

type requestLinkRequest struct {
	Address string `json:"address"`
}

type requestLinkAccepted struct {
	Fingerprint string `json:"fingerprint"`
}

func (l *LinkResponder) waitForLinkAccept(lc *dap.LinkClient, loc *connect.Location, req *Request, addr string) {
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

				modelId, err := mIdentity.CreateIdentityFromDispatch(recvId, "")
				if err != nil {
					logError("[RLT-LR]", "Received error getting model id.", err)
					req.Response <- createError(linkDoneType, "couldn't get id")
					return
				}

				// Load the Server Information
				modelId.Server = loc.Server.Location
				modelId.ServerFingerprint = loc.Server.String()
				modelId.ServerKey = hex.EncodeToString(
					crypto.RSAToBytes(
						loc.Server.EncryptionKey,
					),
				)

				modelId.Nickname = fmt.Sprintf(
					"Imported %s",
					time.Now().Format("3:04PM Jan 2, 2006"),
				)

				// TODO: srv.Alias is likely to be the empty string
				if loc.Server.Alias == "" {
					// Let's get the server alias!
					fmt.Println("WARNING: New linked server has no alias.")
				}

				modelId.ServerAlias = loc.Server.Alias

				// Save the new Identity
				_, err = req.Environment.Tables.Identity.Insert(modelId).
					Exec(req.Environment.Store)
				if err != nil {
					logError("[RLT-LR]", "Received error saving ID.", err)
					req.Response <- createError(linkDoneType, "couldn't save new id")
					return
				}

				// Create and save the new Alias
				addrComponents := strings.Split(addr, "@")
				location := ""
				username := ""
				if len(addrComponents) == 2 {
					location = addrComponents[1]
					username = addrComponents[0]
				} else {
					fmt.Println(addr, "isn't in the correct form for linking.")
				}

				_, err = req.Environment.Tables.Alias.Insert(
					&mIdentity.Alias{
						Identity: gdb.ForeignKey(modelId),
						Location: location,
						Username: username,
					},
				).Exec(req.Environment.Store)
				if err != nil {
					logError("[RLT-LR]", "Received error saving Alias.", err)
					req.Response <- createError(linkDoneType, "couldn't save alias")
					return
				}

				req.Response <- mustCreateMessage(linkDoneType, requestLinkAccepted{
					Fingerprint: modelId.Fingerprint,
				})

				return
			}
		case <-stop.C:
			fmt.Println("Timed out of Receive Identity")
			ticker.Stop()

			req.Response <- createError(linkDoneType, "time out waiting for id")

			return
		}
	}
}

// Requestor
// 1. Send Link Request
// 2. Await Acceptance
// 3. Download Identity

type linkStartResponse struct {
	Code string `json:"code"`
}

func (l *LinkResponder) handleRequestLink(req *Request) {
	msg := req.Message
	obj := &requestLinkRequest{}
	if err := msg.Unmarshal(&obj); err != nil {
		logError("[RLT-LR]", "Received error decoding link request.", err)
		req.Response <- createError(requestedLink, "malformed request")
		return
	}

	// Create Temporary Identity for Router
	tempId, err := identity.CreateIdentity()
	if err != nil {
		logError("[RLT-LR]", "Received error creating identity.", err)
		req.Response <- createError(requestedLink, "identity creation error")
		return
	}

	// Create Router
	c := &connect.Client{
		Origin: tempId,
		Router: &router.Router{
			Origin: tempId,
		},
	}

	loc, err := c.GetLocation(&mIdentity.Address{
		Alias: obj.Address,
	})
	if err != nil {
		logError("[RLT-LR]",
			"Received error looking up address to link (probably not real)", err)
		req.Response <- createError(requestedLink, "lookup error")
		return
	}

	// Create the Link Client
	lc, err := dap.CreateLinkClient(loc.Author, loc.Server)
	if err != nil {
		logError("[RLT-LR]",
			"Received error creating link client.", err)
		req.Response <- createError(requestedLink, "link client error")
		return
	}

	// Create the Verification Code
	vc, err := lc.GetVerificationCode()
	if err != nil {
		logError("[RLT-LR]",
			"Received error making verification code.", err)
		req.Response <- createError(requestedLink, "verification code error")
		return
	}

	// Display the Verification Code
	req.Response <- mustCreateMessage(linkVerificationType, linkStartResponse{
		Code: string(vc),
	})

	// Send the Link Request
	err = lc.LinkRequest(vc)
	if err != nil {
		// This is probably related to the other side not being enabled in time...
		logError("[RLT-LR]", "Received error sending request.", err)
		req.Response <- createError(requestedLink, "couldn't send request")
		return
	}

	go l.waitForLinkAccept(lc, loc, req, obj.Address)

	req.Response <- mustCreateMessage(requestedLink, nil)
}
