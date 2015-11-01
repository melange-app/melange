package realtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"syscall"

	"getmelange.com/backend/models/messages"

	"getmelange.com/backend/framework"
	"getmelange.com/backend/info"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var quitChan = make(chan struct{})

var responders Responders = []Responder{
	&FileResponder{},
	&LinkResponder{},
	&MessageResponder{},
}

// Handler connects one websocket-enabled client to the backend go server.
type Handler struct {
	Environment *info.Environment

	messageChan chan *messages.JSONMessage
	dataChan    chan *Message
	quitChan    chan bool
}

// CreateHandler will build a new object to serve the needs of
// the client connecting.
func CreateHandler(env *info.Environment) *Handler {
	return &Handler{
		Environment: env,

		messageChan: make(chan *messages.JSONMessage),
		dataChan:    make(chan *Message),
		quitChan:    make(chan bool),
	}
}

// getOriginAllowed will determine whether or not an origin connecting
// to the Websocket can be upgraded.
func (r *Handler) getOriginAllowed() func(req *http.Request) bool {
	return func(req *http.Request) bool {
		return req.Header["Origin"][0] == r.Environment.AppURL()
	}
}

// shouldCloseWithError determines whether or not the websocket error
// is grave enough to close the connection.
func shouldCloseWithError(err error) bool {
	return err == io.EOF ||
		err.Error() == "unexpected EOF" ||
		err.Error() == "use of a closed network connection"
}

// UpgradeConnection will upgrade a normal HTTP Request to the Websocket
// protocol.
func (r *Handler) UpgradeConnection(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Requesting a WebSocket Upgrade")
	wsUpgrader.CheckOrigin = r.getOriginAllowed()

	// Update to the websocket protocol
	conn, err := wsUpgrader.Upgrade(res, req, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error", err)

		// Send a 500 on Error
		framework.WriteView(
			&framework.HTTPError{
				ErrorCode: 500,
				Message:   "Websocket upgrade failed.",
			}, res,
		)
		return
	}

	// Attach the websocket to the fetcher.
	r.Environment.Manager.Fetcher.AddObserver(r.messageChan)

	go r.writeMessageLoop(conn)
	go r.readMessageLoop(conn)
}

func (r *Handler) writeMessageLoop(conn *websocket.Conn) {
	for {
		var err error
		select {
		case m := <-r.dataChan:
			err = conn.WriteJSON(m)
		case m := <-r.messageChan:
			err = conn.WriteJSON(map[string]interface{}{
				"type": "message",
				"data": m,
			})
		case <-r.quitChan:
			return
		}

		// Check Write Error
		if err == syscall.EPIPE {
			// Stop!
			return
		} else if err != nil {
			fmt.Println("Error writing to WS", err)
		}
	}
}

func (r *Handler) readMessageLoop(conn *websocket.Conn) {
	for {
		msg := &Message{}

		fmt.Println("Waiting for new message.")
		_, p, err := conn.ReadMessage()
		if err != nil {
			if shouldCloseWithError(err) {
				fmt.Println("Websocket received fatal error", err)
				return
			}

			fmt.Println("Error reading from websocket.", err)
			continue
		}

		err = json.Unmarshal(p, &msg)
		if err != nil {
			fmt.Println("Error unmarshalling from websocket", err)
			continue
		}

		request := &Request{
			Message:     msg,
			Environment: r.Environment,
			Response:    r.dataChan,
		}

		if !responders.Handle(request) {
			fmt.Println("Unable to handle message type", msg.Type)
			continue
		}

		r.dataChan <- msg
	}
}
