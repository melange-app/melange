package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"getmelange.com/zooko/rpc"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
)

// ZookoServer is the object that represents a Zooko Server that
// sends Merkle Branches of Namecoin transactions to users.
type ZookoServer struct {
	// AirDispatch Information
	Key    *identity.Identity
	Router routing.Router

	// Namecoin RPC Information
	*rpc.Server
	Names *NamesManager
}

func (r *ZookoServer) Run(port int) error {
	// Resolve the Address of the Server
	service := ":" + strconv.Itoa(port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)

	// Start the Server
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	r.serverLoop(listener)
	return nil
}

func (r *ZookoServer) serverLoop(listener *net.TCPListener) {
	connections := make(chan net.Conn)

	// Loop forever, waiting for connections
	go func() {
		for {
			// Accept a Connection
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error getting Accept", err)
				continue
			}

			connections <- conn
		}
	}()

	for {
		select {
		case conn := <-connections:
			go r.handleClient(conn)
		}
	}
}

func (r *ZookoServer) handleError(msg string, err error) {
	fmt.Println(msg, err)
}

func (s *ZookoServer) handleClient(conn net.Conn) {
	fmt.Println("Serving", conn.RemoteAddr().String())
	tNow := time.Now()
	defer fmt.Println("Finished with", conn.RemoteAddr().String(), "in", time.Since(tNow).String())

	// Close the Connection after Handling
	defer conn.Close()

	// Read in the Message
	newMessage, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		// There is nothing we can do if we can't read the message.
		s.handleError("Read Message From Connection", err)
		adErrors.CreateError(adErrors.UnexpectedError, "Unable to read message properly.", s.Key.Address).Send(s.Key, conn)
		return
	}

	_, ok := newMessage.Header[s.Key.Address.String()]
	if ok {
		signedMessage, err := newMessage.Decrypt(s.Key)
		if err != nil {
			s.handleError("Decrypt Message", err)
			adErrors.CreateError(adErrors.UnexpectedError, "Unable to decrypt message.", s.Key.Address).Send(s.Key, conn)
			return
		}

		if !signedMessage.Verify() {
			s.handleError("Verify Signature", errors.New("Unable to Verify Signature on Message"))
			adErrors.CreateError(adErrors.InvalidSignature, "Message contains invalid signature.", s.Key.Address).Send(s.Key, conn)
			return
		}

		data, mesType, h, err := signedMessage.ReconstructMessageWithTimestamp()
		if err != nil {
			s.handleError("Verifying Message Structure", err)
			adErrors.CreateError(adErrors.UnexpectedError, "Unable to unpack transfer message.", s.Key.Address).Send(s.Key, conn)
			return
		}

		returnMessage, err := s.handleMessage(data, mesType, h)
		if err != nil {
			fmt.Println("Got error handling message", err)
			returnMessage = adErrors.CreateError(adErrors.UnexpectedError, "Internal service error handling your message", s.Key.Address)
		}

		returnAddress := h.From
		// Lookup from Router if Return Address is not Sendable
		if !h.From.CanSend() {
			if s.Router == nil {
				adErrors.CreateError(adErrors.UnexpectedError, "No router to lookup your address. Must provide return information.", s.Key.Address).Send(s.Key, conn)
				return
			}

			if h.From.Alias != "" {
				// Lookup by Alias
				returnAddress, err = s.Router.LookupAlias(h.From.Alias, routing.LookupTypeDEFAULT)
				if err != nil {
					s.handleError("Looking up Return Address", err)
					adErrors.CreateError(adErrors.UnexpectedError, "Cannot lookup return address.", s.Key.Address).Send(s.Key, conn)
					return
				}
			} else {
				// Lookup by Address
				returnAddress, err = s.Router.Lookup(h.From.String(), routing.LookupTypeDEFAULT)
				if err != nil {
					s.handleError("Looking up Return Address", err)
					adErrors.CreateError(adErrors.UnexpectedError, "Cannot lookup return address.", s.Key.Address).Send(s.Key, conn)
					return
				}
			}
		}

		err = message.SignAndSendToConnection(returnMessage, s.Key, returnAddress, conn)
		if err != nil {
			fmt.Println("Got error sending return message: ", err)
		}
	}
}
