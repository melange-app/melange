package realtime

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"getmelange.com/backend/info"
	mIdentity "getmelange.com/backend/models/identity"
	"getmelange.com/dap"
)

const (
	linkStartedType = "startedLink"
	linkWaitingType = "waitingForLinkRequest"
	linkAcceptType  = "acceptedLink"
)

type linkEnableRequest struct {
	Fingerprint string `json:"fingerprint"`
	ID          string `json:"id"`
}

func (r *LinkResponder) handleStartLink(req *Request) {
	// Unmarshal the request object.
	linkRequest := &linkEnableRequest{}
	err := req.Message.Unmarshal(&linkRequest)
	if err != nil {
		logError("[RLT-LA]", "Received error unmarshalling request.", err)
		req.Response <- createError(linkStartedType, "invalid request")
		return
	}

	// Find the identity to link.
	modelID, err := r.getIdentityWithFingerprint(req.Environment, linkRequest.ID)
	if err != nil {
		logError("[RLT-LA]", "Received error getting identity to enable.", err)
		req.Response <- createError(linkStartedType, "couldn't get identity")
		return
	}

	// Create the DAP Client
	client, err := modelID.Client(req.Environment.Store)
	if err != nil {
		logError("[RLT-LA]", "Received error getting DAP client", err)
		req.Response <- createError(linkStartedType, "couldn't get DAP client")
		return
	}

	err = client.EnableLink()
	if err != nil {
		logError("[RLT-LA]", "Received error enabling link", err)
		req.Response <- createError(linkStartedType, "couldn't enable link")
		return
	}

	req.Response <- mustCreateMessage(linkWaitingType, nil)

	go r.waitForLinkRequest(client, req)
}

type linkReceivedResponse struct {
	Code string `json:"code"`
	UUID string `json:"uuid"`
}

func (r *LinkResponder) waitForLinkRequest(client *dap.Client, req *Request) {
	request, err := client.WaitForLinkRequest()
	if err != nil {
		logError("[RLT-LA]", "Received error waiting for request", err)
		req.Response <- createError(linkStartedType, "couldn't wait for link")
	}

	// Create the random string to use as our identifier.
	id := uuid.New()

	req.Response <- mustCreateMessage(linkStartedType, linkReceivedResponse{
		Code: string(request.Verification),
		UUID: id,
	})

	// Include the pending link request.
	r.requestsLock.Lock()
	r.requests[id] = &ongoingLink{
		PendingLinkRequest: request,
		Client:             client,
	}
	r.requestsLock.Unlock()
}

type linkAcceptRequest struct {
	UUID string `json:"uuid"`
}

func (r *LinkResponder) handleAcceptLink(req *Request) {
	r.requestsLock.RLock()
	defer r.requestsLock.RUnlock()

	// Unmarshal the linkAcceptRequest
	data := &linkAcceptRequest{}
	err := req.Message.Unmarshal(&data)
	if err != nil {
		logError("[RLT-LA]", "Received error unmarshalling accept request.", err)
		req.Response <- createError(linkAcceptType, "invalid request")
		return
	}

	// Find the pending request
	request, ok := r.requests[data.UUID]
	if !ok {
		logError("[RLT-LA]", fmt.Sprintf("UUID returned (%s) does not exist.", data.UUID))
		req.Response <- createError(linkAcceptType, "invalid request - no such uuid")
		return
	}

	// Accept the link request
	if err := request.accept(); err != nil {
		logError("[RLT-LA]", "Received error accepting link request.", err)
		req.Response <- createError(linkAcceptType, "couldn't accept link request")
		return
	}

	req.Response <- mustCreateMessage("acceptedLink", nil)
}

func (r *LinkResponder) getIdentityWithFingerprint(env *info.Environment, fp string) (*mIdentity.Identity, error) {
	// Grab the Identity from the Table
	obj := &mIdentity.Identity{}
	err := env.Tables.Identity.Get().Where("fingerprint", fp).One(env.Store, obj)

	return obj, err
}
