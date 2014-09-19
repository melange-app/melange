package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"getmelange.com/app/framework"
	"getmelange.com/app/models"
	"github.com/gorilla/websocket"
	gdb "github.com/huntaub/go-db"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header["Origin"][0] == "http://app.melange:7776"
	},
}
var quitChan = make(chan struct{})

// RealtimeHandler connects one websocket-enabled client to the backend go server.
type RealtimeHandler struct {
	Store  *models.Store
	Tables map[string]gdb.Table

	messageChan chan *models.JSONMessage
	dataChan    chan interface{}
}

// UpgradeConnection will upgrade a normal HTTP Request to the Websocket
// protocol.
func (r *RealtimeHandler) UpgradeConnection(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Requesting a WebSocket Upgrade")
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

	r.dataChan = make(chan interface{})
	r.messageChan = make(chan *models.JSONMessage)

	// Attach the websocket to the fetcher.
	models.AddFetchWatcher(r.messageChan)

	go func(mes <-chan *models.JSONMessage, data <-chan interface{}, quit <-chan struct{}) {
		for {
			select {
			case m := <-data:
				err := conn.WriteJSON(m)
				if err != nil {
					fmt.Println("Error writing to WS", err)
				}
			case m := <-mes:
				err := conn.WriteJSON(map[string]interface{}{
					"type": "message",
					"data": m,
				})
				if err != nil {
					fmt.Println("Error writing to WS", err)
				}
			case <-quit:
				return
			}
		}
	}(r.messageChan, r.dataChan, quitChan)

	go func(conn *websocket.Conn, dataChan chan interface{}) {
		for {
			v := make(map[string]interface{})

			fmt.Println("Waiting for new message.")
			_, p, err := conn.ReadMessage()
			if err != nil {
				if err == io.EOF {
					fmt.Println("Lost connection from websocket.")
					return
				} else if err.Error() == "unexpected EOF" {
					fmt.Println("Lost connection from websocket.")
					return
				}

				fmt.Println("Error reading from websocket.", err)
				continue
			}

			err = json.Unmarshal(p, &v)
			if err != nil {
				fmt.Println("Error unmarshalling from websocket", err)
				continue
			}

			typ, ok1 := v["type"]
			data, ok2 := v["data"]
			if !ok1 || !ok2 {
				fmt.Println("Websocket doesn't have correct message type.")
				continue
			}

			typString, ok := typ.(string)
			if !ok {
				fmt.Println("Websocket message type has incorrect format.")
				continue
			}

			outTyp, outData := r.HandleWSRequest(typString, data)

			dataChan <- map[string]interface{}{
				"type": outTyp,
				"data": outData,
			}
		}
	}(conn, r.dataChan)
}

// HandleWSRequest will handle one WS request from the client.
func (r *RealtimeHandler) HandleWSRequest(t string, d interface{}) (string, interface{}) {
	if t == "startup" {
		// We need to send all current messages to the client. Initiate transfer.
		m, err := constructManager(r.Store, r.Tables)
		if err != nil {
			fmt.Println("Unable to construct WS Message Manager", err)
		}

		m.GetAllMessages(r.messageChan)
		return "initDone", nil
	}
	fmt.Println("Got message", t, "from ws.")
	return "gotIt", nil
}
