package realtime

import "getmelange.com/backend/models/messages"

// MessageResponder will return all of messages that are currently in
// the system.
type MessageResponder struct{}

// Handle will perform the retrieval.
func (m *MessageResponder) Handle(req *Request) bool {
	if req.Message.Type == "startup" {
		msgChan := make(chan *messages.JSONMessage)

		go func() {
			for {
				msg, ok := <-msgChan
				if !ok {
					return
				}

				req.Response <- mustCreateMessage("message", msg)
			}
		}()

		go func() {
			req.Environment.Manager.GetAllMessages(msgChan)
			req.Response <- mustCreateMessage("initDone", nil)
		}()

		return true
	}

	return false
}
