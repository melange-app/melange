package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"getmelange.com/zooko/message"
	"getmelange.com/zooko/rpc"

	"github.com/melange-app/nmcd/btcjson"
)

const (
	nameNewConfirmations         = 12
	nameFirstUpdateConfirmations = 1

	confirmationTime      = 10 * time.Minute
	checkConfirmationTime = 5 * time.Minute
)

type lookupRequest struct {
	Name     string
	Response chan *[]byte
}

type registerRequest struct {
	Name  string
	Value []byte

	// TxID is the transaction identifier of the name_new
	// transaction. We will watch this for updates and send
	// name_firstupdate when we need to.
	TxID            string
	NameFirstUpdate []byte

	// This is how we keep track of transaction that will not
	// actually make it into the blockchain and purge them from
	// the cache if necessary.
	Ticks    int
	Attempts int
}

// NamesManager is a system that keeps track of all of the Namecoin
// registrations currently waiting to be processed.
type NamesManager struct {
	*rpc.Server

	// Storage for the cached names.
	cached map[string]*registerRequest

	// Actions that we can receive in the loop.
	lookup            chan lookupRequest
	register          chan *registerRequest
	confirmationTimer *time.Ticker
}

// CreateNamesManager will create a new Name Management system that
// interfaces with Namecoin and implements a cache for when names
// aren't yet in the blockchain.
func CreateNamesManager(s *rpc.Server) *NamesManager {
	n := &NamesManager{
		Server: s,

		// Make the maps and channels
		cached:   make(map[string]*registerRequest),
		lookup:   make(chan lookupRequest),
		register: make(chan *registerRequest),

		// Make the timer
		confirmationTimer: time.NewTicker(checkConfirmationTime),
	}

	// Start the channel looking loop.
	go n.loop()

	return n
}

func (n *NamesManager) loop() {
	for {
		select {
		case l := <-n.lookup:
			val, ok := n.cached[l.Name]
			if !ok {
				l.Response <- nil
				continue
			}

			l.Response <- &val.Value

		case r := <-n.register:
			n.cached[r.Name] = r
		case <-n.confirmationTimer.C:
			n.checkForConfirmations()
		}
	}
}

func (n *NamesManager) checkForConfirmations() {
	newCache := make(map[string]*registerRequest)

	for name, reg := range n.cached {
		confirmations, err := n.Server.Confirmations(reg.TxID)
		if err != nil {
			fmt.Println(
				"Got error checking for confirmation on",
				name, err)
			continue
		}

		if reg.NameFirstUpdate == nil &&
			confirmations >= nameFirstUpdateConfirmations {
			// In this situation, we can simply remove the
			// transaction from the cache.
		} else if confirmations >= nameNewConfirmations {
			// In this situation, we need to broadcast the
			// new transaction and get the txid of it to
			// place back in the cache and wait more... :/
			rawTx := hex.EncodeToString(reg.NameFirstUpdate)
			txId, err := n.broadcastAndGetID(rawTx)
			if err != nil {
				// If we have consistently gotten an
				// error on the past 5 attempts, we
				// will not continue taking up space
				// in the cache.
				if reg.Attempts > 5 {
					fmt.Println("[ZOOKO] Removing", name, "from the cache for getting 5 successive errors on broadcast transaction.")
					continue
				}

				fmt.Println("[ZOOKO] Received error while broadcast name_firstupdate for", name, err)

				// If we get an error at this point,
				// we are pretty screwed. We are going
				// to increase a marker, and we will
				// remove from the cache on too many
				// attempts.
				reg.Ticks++
				reg.Attempts++
				newCache[name] = reg

				continue
			}

			// Input the new request.
			reg.Ticks = 0
			reg.TxID = txId
			newCache[name] = reg
		} else {
			// Otherwise we populate the new cache with
			// the name immediately.
			reg.Ticks++

			// In this situation, we have waited a LONG
			// time to get a confirmation, we should
			// assume that the network will not accept
			// this transaction and remove it from the
			// chain.
			if time.Duration(reg.Ticks)*checkConfirmationTime > (2*nameNewConfirmations*confirmationTime) &&
				confirmations == 0 {
				fmt.Println("[ZOOKO] Removing", name, "from the cache for failing to get a confirmation.")
				continue
			}

			newCache[name] = reg
		}
	}

	// Overwrite the cache.
	if len(newCache) != 0 {
		n.cached = newCache
	}
}

func (n *NamesManager) broadcastAndGetID(rawTx string) (string, error) {
	// Broadcast the Transaction
	if err := n.Server.Broadcast(rawTx); err != nil {
		return "", err
	}

	// Get the TxID from the raw transaction
	cmd, err := btcjson.NewDecodeRawTransactionCmd(nil, rawTx)
	if err != nil {
		return "", err
	}

	result, err := n.Server.Send(cmd)
	if err != nil {
		return "", err
	} else if result.Error != nil {
		return "", *result.Error
	}

	// Get the Transaction ID
	return result.Result.(*btcjson.TxRawDecodeResult).Txid, nil
}

func (n *NamesManager) Register(msg *message.RegisterName) (bool, error) {
	if _, ok, err := n.Lookup(*msg.Name); err != nil || ok {
		return false, err
	}

	rawTx := hex.EncodeToString(msg.NameNew)
	txId, err := n.broadcastAndGetID(rawTx)
	if err != nil {
		return true, err
	}

	n.register <- &registerRequest{
		Name:            *msg.Name,
		Value:           msg.Value,
		TxID:            txId,
		NameFirstUpdate: msg.NameFirstupdate,
	}
	return true, nil
}

func (n *NamesManager) Renew(msg *message.RenewName) error {
	if _, ok, err := n.Lookup(*msg.Name); err != nil || !ok {
		return errors.New("zooko/server: name lookup failed before renew")
	}

	rawTx := hex.EncodeToString(msg.NameUpdate)
	txId, err := n.broadcastAndGetID(rawTx)
	if err != nil {
		return err
	}

	n.register <- &registerRequest{
		Name:  *msg.Name,
		Value: msg.Value,
		TxID:  txId,
	}
	return nil
}

func (n *NamesManager) Lookup(name string) ([]byte, bool, error) {
	resp := make(chan *[]byte)
	n.lookup <- lookupRequest{
		Name:     name,
		Response: resp,
	}

	data := <-resp
	if data != nil {
		return *data, true, nil
	}

	val, found, err := n.Server.LookupName(name)
	if err != nil || !found {
		return nil, false, err
	}

	return []byte(val), true, nil
}
