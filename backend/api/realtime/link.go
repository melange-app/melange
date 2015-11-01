package realtime

import (
	"sync"

	"getmelange.com/dap"
)

type ongoingLink struct {
	*dap.PendingLinkRequest
	Client *dap.Client
}

func (o *ongoingLink) accept() error {
	return o.Client.LinkAcceptRequest(o.PendingLinkRequest)
}

// LinkResponder
type LinkResponder struct {
	requests     map[string]*ongoingLink
	requestsLock *sync.RWMutex
}

func CreateLinkResponder() *LinkResponder {
	return &LinkResponder{
		requestsLock: &sync.RWMutex{},
		requests:     make(map[string]*ongoingLink),
	}
}

func (l *LinkResponder) Handle(req *Request) bool {
	if req.Message.Type == "requestLink" {
		go l.handleRequestLink(req)
		return true
	} else if req.Message.Type == "startLink" {
		go l.handleStartLink(req)
		return true
	} else if req.Message.Type == "acceptLink" {
		go l.handleAcceptLink(req)
		return true
	}

	return false
}
