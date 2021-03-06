package dap

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"strings"
	"testing"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/server"
	adTest "airdispat.ch/testing"
	adWire "airdispat.ch/wire"
)

func testingSetup(t *testing.T) (quit chan bool, results chan *TestingResult, scene adTest.Scenario, client *Client) {
	scene, err := adTest.CreateScenario()
	if err != nil {
		t.Error(err.Error())
		return
	}

	results = make(chan *TestingResult, 1)
	delegate := &TestingDelegate{
		Results: results,
	}

	started := make(chan bool)
	quit = make(chan bool)

	theServer := server.Server{
		LocationName: "localhost:9091",
		Key:          scene.Server,
		Delegate:     &server.BasicServer{},
		Router:       scene.Router,
		Handlers: []server.Handler{
			&Handler{
				Key:      scene.Server,
				Delegate: delegate,
			},
		},
		Start: started,
		Quit:  quit,
	}

	go func() {
		theServer.StartServer("9091")
	}()

	delegate.Scenario = scene

	client = &Client{
		Key:    scene.Sender,
		Server: scene.Server.Address,
	}

	// Wait for Started, begin Tests
	<-started

	return
}

// Test Delegate Registration Method
func TestRegister(t *testing.T) {
	quit, results, _, client := testingSetup(t)
	defer func() { quit <- true }()

	err := client.Register(map[string][]byte{
		"full_name": []byte("John Smith"),
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

// Test Delegate Unregister Method
func TestUnregister(t *testing.T) {
	quit, results, _, client := testingSetup(t)
	defer func() { quit <- true }()

	err := client.Unregister(map[string][]byte{
		"reason": []byte("Too expensive!"),
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

// Test GetMessages Method
func TestGetMessages(t *testing.T) {
	quit, results, scene, client := testingSetup(t)
	defer func() { quit <- true }()

	response, err := client.DownloadMessages(1, true)
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
		return
	}

	if len(response) != 1 {
		t.Error("Incorrect number of messages returned.")
		return
	}
	res := response[0]

	readStatus, ok := res.Context["read"]
	if !ok {
		t.Error("Cannot find read status of message.")
		return
	}
	if string(readStatus) != "yes" {
		t.Error("Incorrect read status of message.")
		return
	}

	receivedSign, err := res.Message.Decrypt(scene.Receiver)
	if err != nil {
		t.Error("Decrypting Message: " + err.Error())
		return
	}

	if !receivedSign.Verify() {
		t.Error("Verifying Signature: " + err.Error())
		return
	}

	data, typ, h, err := receivedSign.ReconstructMessageWithTimestamp()
	if err != nil {
		t.Error("Reconstructing Message: " + err.Error())
		return
	}

	if typ != adWire.MailCode {
		t.Error("Type of Message is unexpected, " + typ)
		return
	}

	mail, err := message.CreateMailFromBytes(data, h)
	if err != nil {
		t.Error("Creating Mail Message: " + err.Error())
		return
	}

	if !mail.Components.HasComponent("test_verification") {
		t.Error("Message doesn't have right components.")
		return
	}

	if mail.Components.GetStringComponent("test_verification") != "Pass" {
		t.Error("Message test_verification doesn't match.")
		return
	}
}

// Test Delegate Publish Method
func TestPublish(t *testing.T) {
	quit, results, scene, client := testingSetup(t)
	defer func() { quit <- true }()

	mail := message.CreateMail(scene.Sender.Address, time.Now(), "test_message", scene.Receiver.Address)
	mail.Components.AddComponent(
		message.Component{
			Name: "test_verification",
			Data: []byte("Pass"),
		},
	)

	_, err := client.PublishMessage(mail, []*identity.Address{
		scene.Receiver.Address,
	}, "test_message", true)
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

// Test Delegate Update Message
func TestUpdate(t *testing.T) {
	quit, results, scene, client := testingSetup(t)
	defer func() { quit <- true }()

	mail := message.CreateMail(scene.Sender.Address, time.Now(), "test_message", scene.Receiver.Address)
	mail.Components.AddComponent(
		message.Component{
			Name: "test_verification",
			Data: []byte("Pass"),
		},
	)

	err := client.UpdateMessage(mail, []*identity.Address{scene.Receiver.Address}, "test_message")
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

// Test Delegate Update Message
func TestPublishDataMessage(t *testing.T) {
	quit, results, scene, client := testingSetup(t)
	defer func() { quit <- true }()

	data := bytes.NewReader([]byte("hello-world"))

	_, err := client.PublishDataMessage(
		data,
		[]*identity.Address{scene.Receiver.Address},
		"airdispat.ch/data",
		"helloData",
		"helloData.txt",
	)

	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

func TestGetData(t *testing.T) {
	quit, results, _, client := testingSetup(t)
	defer func() { quit <- true }()

	data, err := client.GetData("test_data")
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}

	if string(data) != "Pass" {
		t.Error("Returned data is incorrect.")
	}
}

func TestSetData(t *testing.T) {
	quit, results, _, client := testingSetup(t)
	defer func() { quit <- true }()

	err := client.SetData("test_data", []byte("Pass"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	result := <-results
	if result != nil {
		t.Error(result.Error())
	}
}

func TestLinking(t *testing.T) {
	quit, _, scene, client := testingSetup(t)
	defer func() { quit <- true }()

	lc, err := CreateLinkClient(scene.Sender.Address, scene.Server.Address)
	if err != nil {
		t.Error(err.Error())
	}

	err = client.EnableLink()
	if err != nil {
		t.Error(err.Error())
	}

	vc, err := lc.GetVerificationCode()
	if err != nil {
		t.Error(err.Error())
	}

	err = lc.LinkRequest(vc)
	if err != nil {
		t.Error(err.Error())
	}

	request, err := client.GetLinkRequest()
	if err != nil {
		t.Error(err.Error())
	}

	if request.From.String() != lc.Client.Key.Address.String() {
		t.Error("Identity mismatch.")
	}

	if request.Verification != string(vc) {
		t.Error("Verification mismatch.")
	}

	err = client.LinkAcceptRequest(request)
	if err != nil {
		t.Error(err.Error())
	}

	recvId, err := lc.LinkGetIdentity()
	if err != nil {
		t.Error(err.Error())
	}

	if recvId.Address.String() != scene.Sender.Address.String() {
		t.Error("Final id mismatch.")
	}
}

type TestingResult struct {
	Location string
	Err      string
}

func (t *TestingResult) IsError() bool {
	return t.Err != ""
}

func (t *TestingResult) Error() string {
	return fmt.Sprintf("%s: %s", t.Location, t.Err)
}

type TestingDelegate struct {
	Scenario adTest.Scenario
	Results  chan *TestingResult
}

// Mock Registration Method
func (t *TestingDelegate) Register(addr string, keys map[string][]byte) error {
	if addr != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Register", "Checking address is correct."}
		return nil
	}

	name, ok := keys["full_name"]
	if !ok || string(name) != "John Smith" {
		t.Results <- &TestingResult{"Register", "Checking dictionary is correct."}
		return nil
	}

	t.Results <- nil
	return nil
}

// Mock Unregistration Method
func (t *TestingDelegate) Unregister(addr string, keys map[string][]byte) error {
	if addr != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Unregister", "Checking address is correct."}
		return nil
	}

	name, ok := keys["reason"]
	if !ok || string(name) != "Too expensive!" {
		t.Results <- &TestingResult{"Unregister", "Checking dictionary is correct."}
		return nil
	}

	t.Results <- nil
	return nil
}

func (t *TestingDelegate) GetSentMessages(since uint64, owner string, context bool) ([]*ResponseMessage, error) {
	panic("Shouldn't be calling GetSentMessages.")
	return nil, nil
}

// Mock GetMessages
func (t *TestingDelegate) GetMessages(since uint64, owner string, context bool) ([]*ResponseMessage, error) {
	if owner != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"GetMessages", "Checking address is correct."}
		return nil, nil
	}

	if since != 1 {
		t.Results <- &TestingResult{"GetMessages", "Checking that since value is correct."}
		return nil, nil
	}

	if !context {
		t.Results <- &TestingResult{"GetMessages", "Checking that context value is correct."}
		return nil, nil
	}

	mail := message.CreateMail(t.Scenario.Sender.Address, time.Now(), "test_message", t.Scenario.Receiver.Address)
	mail.Components.AddComponent(
		message.Component{
			Name: "test_verification",
			Data: []byte("Pass"),
		},
	)

	signed, err := message.SignMessage(mail, t.Scenario.Sender)
	if err != nil {
		t.Results <- &TestingResult{"GetMessages, Signing Message", err.Error()}
		return nil, nil
	}

	encrypted, err := signed.EncryptWithKey(t.Scenario.Receiver.Address)
	if err != nil {
		t.Results <- &TestingResult{"GetMessages, Encrypting Message", err.Error()}
		return nil, nil
	}

	resMsg := CreateResponseMessage(encrypted, t.Scenario.Server.Address, t.Scenario.Sender.Address)
	resMsg.Context["read"] = []byte("yes")

	t.Results <- nil
	return []*ResponseMessage{resMsg}, nil
}

// Mock PublishMessge
func (t *TestingDelegate) PublishMessage(name string, to []string, author string, msg *message.EncryptedMessage, alerted bool) error {
	// Verifiy Arguments
	if author != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Publish Message", "Checking address is correct."}
		return nil
	}

	if name != "test_message" {
		t.Results <- &TestingResult{"Publish Message", "Checking that name is correct."}
		return nil
	}

	if len(to) != 1 || to[0] != t.Scenario.Receiver.Address.String() {
		t.Results <- &TestingResult{"Publish Message", "Checking that to is correct."}
		return nil
	}

	if !alerted {
		t.Results <- &TestingResult{"Publish Message", "Checking that alerted is correct."}
		return nil
	}

	// Verify Message
	receivedSign, err := msg.Decrypt(t.Scenario.Receiver)
	if err != nil {
		t.Results <- &TestingResult{"Publish Message, Decryption", err.Error()}
		return nil
	}

	if !receivedSign.Verify() {
		t.Results <- &TestingResult{"Publish Message, Verification", err.Error()}
		return nil
	}

	data, typ, h, err := receivedSign.ReconstructMessageWithTimestamp()
	if err != nil {
		t.Results <- &TestingResult{"Publish Message, Reconstruction", err.Error()}
		return nil
	}

	if typ != adWire.MailCode {
		t.Results <- &TestingResult{"Publish Message", "Type of Message is unexpected, " + typ}
		return nil
	}

	mail, err := message.CreateMailFromBytes(data, h)
	if err != nil {
		t.Results <- &TestingResult{"Publish Message, Creation", err.Error()}
		return nil
	}

	if !mail.Components.HasComponent("test_verification") {
		t.Results <- &TestingResult{"Publish Message", "Message doesn't have right components."}
		return nil
	}

	if mail.Components.GetStringComponent("test_verification") != "Pass" {
		t.Results <- &TestingResult{"Publish Message", "Message test_verification doesn't match."}
		return nil
	}

	t.Results <- nil
	return nil
}

func (t *TestingDelegate) PublishDataMessage(name string, to []string, author string, msg *message.EncryptedMessage, length uint64, r ReadVerifier) error {
	// Verifiy Arguments
	if author != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Publish Data Message", "Checking address is correct."}
		return nil
	}

	if !strings.HasPrefix(name, "helloData") {
		t.Results <- &TestingResult{"Publish Data Message", "Checking that name is correct."}
		return nil
	}

	if len(to) != 1 || to[0] != t.Scenario.Receiver.Address.String() {
		t.Results <- &TestingResult{"Publish Data Message", "Checking that to is correct."}
		return nil
	}

	if length != 11+aes.BlockSize {
		t.Results <- &TestingResult{"Publish Data Message", "Checking that length is correct."}
		return nil
	}

	// Verify Message
	receivedSign, err := msg.Decrypt(t.Scenario.Receiver)
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Decryption", err.Error()}
		return nil
	}

	if !receivedSign.Verify() {
		t.Results <- &TestingResult{"Publish Data Message, Verification", err.Error()}
		return nil
	}

	data, typ, h, err := receivedSign.ReconstructMessageWithTimestamp()
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Reconstruction", err.Error()}
		return nil
	}

	if typ != adWire.DataCode {
		t.Results <- &TestingResult{"Publish Data Message", "Type of Message is unexpected, " + typ}
		return nil
	}

	mail, err := message.CreateDataMessageFromBytes(data, h)
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Creation", err.Error()}
		return nil
	}

	b := &bytes.Buffer{}
	_, err = b.ReadFrom(r)
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Reading from Verifier", err.Error()}
		return nil
	}

	// Verify the Encrypted Payload
	if !r.Verify() {
		t.Results <- &TestingResult{"Publish Data Message", "Verification of payload."}
		return nil
	}

	if uint64(b.Len()) != mail.Length {
		t.Results <- &TestingResult{"Publish Data Message", "Length mistmatch."}
		return nil
	}

	// Time to decrypt the plaintext.
	d := &bytes.Buffer{}
	decrypted, err := mail.DecryptReader(b)
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Creating Decrypter", err.Error()}
		return nil
	}

	_, err = d.ReadFrom(decrypted)
	if err != nil {
		t.Results <- &TestingResult{"Publish Data Message, Reading from Decrypter", err.Error()}
		return nil
	}

	if !mail.VerifyPayload() {
		t.Results <- &TestingResult{"Publish Data Message", "Unable to verify plaintext."}
		return nil
	}

	if !bytes.Equal(d.Bytes(), []byte("hello-world")) {
		t.Results <- &TestingResult{"Publish Data Message", "Payload is incorrect."}
		return nil
	}

	t.Results <- nil
	return nil
}

// Mock UpdateMessage
func (t *TestingDelegate) UpdateMessage(name string, author string, msg *message.EncryptedMessage) error {
	// Verifiy Arguments
	if author != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Update Message", "Checking address is correct."}
		return nil
	}

	if name != "test_message" {
		t.Results <- &TestingResult{"Update Message", "Checking that name is correct."}
		return nil
	}

	// Verify Message
	receivedSign, err := msg.Decrypt(t.Scenario.Receiver)
	if err != nil {
		t.Results <- &TestingResult{"Update Message, Decryption", err.Error()}
		return nil
	}

	if !receivedSign.Verify() {
		t.Results <- &TestingResult{"Update Message, Verification", err.Error()}
		return nil
	}

	data, typ, h, err := receivedSign.ReconstructMessageWithTimestamp()
	if err != nil {
		t.Results <- &TestingResult{"Update Message, Reconstruction", err.Error()}
		return nil
	}

	if typ != adWire.MailCode {
		t.Results <- &TestingResult{"Update Message", "Type of Message is unexpected, " + typ}
		return nil
	}

	mail, err := message.CreateMailFromBytes(data, h)
	if err != nil {
		t.Results <- &TestingResult{"Update Message, Creation", err.Error()}
		return nil
	}

	if !mail.Components.HasComponent("test_verification") {
		t.Results <- &TestingResult{"Update Message", "Message doesn't have right components."}
		return nil
	}

	if mail.Components.GetStringComponent("test_verification") != "Pass" {
		t.Results <- &TestingResult{"Update Message", "Message test_verification doesn't match."}
		return nil
	}

	t.Results <- nil
	return nil
}

// Mock GetData
func (t *TestingDelegate) GetData(owner string, key string) ([]byte, error) {
	// Verifiy Arguments
	if owner != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Get Data", "Checking address is correct."}
		return nil, nil
	}

	if key != "test_data" {
		t.Results <- &TestingResult{"Get Data", "Checking that name is correct."}
		return nil, nil
	}

	t.Results <- nil
	return []byte("Pass"), nil
}

// Mock SetData
func (t *TestingDelegate) SetData(owner string, key string, data []byte) error {
	// Verifiy Arguments
	if owner != t.Scenario.Sender.Address.String() {
		t.Results <- &TestingResult{"Get Data", "Checking address is correct."}
		return nil
	}

	if key != "test_data" {
		t.Results <- &TestingResult{"Get Data", "Checking that name is correct."}
		return nil
	}

	if string(data) != "Pass" {
		t.Results <- &TestingResult{"Get Data", "Check that data is correct."}
		return nil
	}

	t.Results <- nil
	return nil
}
