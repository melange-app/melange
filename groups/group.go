package groups

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/wire"
	"code.google.com/p/go-uuid/uuid"
	"getmelange.com/dap"
)

// Group Messages
// --------------
// key      - current group symmetric key
// members  - current list of members

type Group struct {
	// Group Keys
	oldKeys []*identity.Identity

	// Group Identity
	identity *identity.Identity
	name     string
	uniqueID string

	// Server Information
	server *identity.Address

	// A list of members in this group.
	members []string

	router routing.Router
}

const (
	groupInvite         string = "group-invite"
	groupInviteIdentity string = "group-invite-identity"
	groupInviteServer   string = "group-invite-server"
	groupInviteMembers  string = "group-invite-members"
	groupInviteID       string = "group-invite-unique-id"
	groupRekeyIdentity  string = "group-rekey-identity"
)

// CreateGroup will create a new Melange group on a server.
//
// It performs the following sequence of actions:
// 1. Generate a new identity.
// 2. Register the identity with a server.
// 3. Choose a random symmetric key to use for group messages.
// 4. Send invitations to each original member:
//    - Signed by the group creator
//    - Signed by the group identity
//    - Containing the group identity
//    - Containing the symmetric key
func CreateGroup(
	// The Human Readable Name of the Group (e.g. 'Lunch Group').
	name string,
	// The "Owner" of the Group, i.e. the person who is sending the group invitations.
	owner *identity.Identity,
	// The server that will host the group content.
	server *identity.Address,
	// The members of the initial group who will be invited.
	r routing.Router,
	memberString ...string,
) (*Group, error) {
	if server.Location == "" {
		panic("You cannot create a group without having a location-aware server.")
	}

	id, err := identity.CreateIdentity()
	if err != nil {
		return nil, err
	}

	// Create a DAP Client with the ID
	client := dap.CreateClient(id, server)

	// Register the Identity with the Server
	if err := client.Register(map[string][]byte{}); err != nil {
		return nil, err
	}

	// groupId is a random uuid that is assigned to this group to
	// identify its messages.
	groupID := uuid.New()

	// Create the Group Object
	g := &Group{
		identity: id,
		name:     name,
		server:   server,
		members:  memberString,
		router:   r,
		uniqueID: groupID,
	}

	members, err := g.loadMembers()
	if err != nil {
		return nil, err
	}

	// Invite the new members.
	return g, g.inviteMembers(owner, members...)
}

// CreateGroupFromInvitation will create a Group object for use.
func CreateGroupFromInvitation(invite *message.Mail, router routing.Router) (*Group, error) {
	c := invite.Components

	gID, err := identity.GobDecodeKey(bytes.NewBuffer(c.GetComponent(groupInviteID)))
	if err != nil {
		return nil, err
	}

	gSrvByt := c.GetComponent(groupInviteServer)
	gSrv, err := identity.DecodeAddress(gSrvByt)
	if err != nil {
		return nil, err
	}

	gMembers := c.GetStringComponent(groupInviteMembers)

	return &Group{
		name:     c.GetStringComponent(groupInvite),
		server:   gSrv,
		identity: gID,
		members:  strings.Split(gMembers, ","),
		router:   router,
		uniqueID: c.GetStringComponent(groupInviteID),
	}, nil
}

func (g *Group) inviteMembers(owner *identity.Identity, newMembers ...*identity.Address) error {
	inviteID := uuid.New()
	inviteName := fmt.Sprintf("group-invitation/%s/%s", g.uniqueID, inviteID)

	// Send Invitations
	invite := message.CreateMail(
		owner.Address,
		time.Now(),
		inviteName,
		newMembers...,
	)

	// Encode the Identity of the Group
	idBuffer := &bytes.Buffer{}
	_, err := g.identity.GobEncodeKey(idBuffer)
	if err != nil {
		return err
	}

	// Encode the Full Address of the Server
	srvBytes, err := g.server.Encode()
	if err != nil {
		return err
	}

	invite.Components = message.ComponentList{
		groupInvite:         message.CreateStringComponent(groupInvite, g.name),
		groupInviteIdentity: message.CreateComponent(groupInviteIdentity, idBuffer.Bytes()),
		groupInviteServer:   message.CreateComponent(groupInviteServer, srvBytes),
		groupInviteMembers:  message.CreateStringComponent(groupInviteMembers, strings.Join(g.members, ",")),
		groupInviteID:       message.CreateStringComponent(groupInviteID, g.uniqueID),
	}

	// Send out the invitation message.
	_, err = g.client().PublishMessage(
		invite,
		newMembers,
		inviteName,
		true,
	)
	return err
}

func (g *Group) loadMembers() ([]*identity.Address, error) {
	var members []*identity.Address
	for _, v := range g.members {
		addr, err := g.router.LookupAlias(v, routing.LookupTypeMAIL)
		if err != nil {
			return nil, err
		}

		members = append(members, addr)
	}

	return members, nil
}

func (g *Group) client() *dap.Client {
	return dap.CreateClient(g.identity, g.server)
}

// Delete will deregister a group (identity) from a server
// thus deleting all of its messages.
func (g *Group) Delete() error {
	return g.client().Unregister(map[string][]byte{})
}

// Group Message Operations

// GetMessages will download all messages from a group
// in order to display them to the user. Remember that all
// group communication is considered "public" so we will utilize
// the RetrievePublicMail method.
func (g *Group) GetMessages(since uint64) ([]*message.Mail, error) {
	rsp, err := g.client().DownloadSentMessages(since, false)
	if err != nil {
		return nil, nil
	}

	var out []*message.Mail
	for _, v := range rsp {
		// TODO: some mechanism to cycle through identities on membership change
		data, typ, h, err := v.Message.Reconstruct(g.identity, false)
		if err != nil {
			fmt.Println("Error reconstructing group message", err)
			continue
		}

		if typ != wire.MailCode {
			fmt.Println("Incorrect type for group message. Expected", wire.MailCode, "got", typ)
			continue
		}

		m, err := message.CreateMailFromBytes(data, h)
		if err != nil {
			fmt.Println("Error creating mail from bytes", err)
			continue
		}

		out = append(out, m)
	}
	return out, nil
}

// PostMessage will post an AirDispatch message to a
// specific Melange group. Specifically, it will do the following.
//
// 1. Sign the message with the keys of the group and the author.
// 2. Encrypt the message with the current group identity / key.
// 3. Publish the message as "public" from the group identity.
func (g *Group) PostMessage(msg *message.Mail, sender *identity.Identity) (string, error) {
	// Sign Message and Encrypt
	sm, err := message.SignMessage(msg, sender)
	if err != nil {
		return "", err
	}

	err = sm.AddSignature(g.identity)
	if err != nil {
		return "", err
	}

	enc, err := sm.EncryptWithKey(g.identity.Address)
	if err != nil {
		return "", err
	}

	// Publish the Signed Message
	return g.client().PublishEncryptedMessage(
		enc,
		[]*identity.Address{g.identity.Address},
		"",
		false,
	)
}

// RemoveMessageFromGroup will remove an AirDispatch message
// from a group.
func (g *Group) RemoveMessage(name string) error {
	// Currently no method exists to delete an AirDispatch message from a server...
	// g.client().
	return nil
}

// Group Management Operations
//
// These operations are unique in that they change how
// the group is constructed unilaterally. Therefore,
// other clients in the group have the ability to "agree"
// with the change proposed by one user, or "deny" it.

// rekey will change the group identity EncryptionKey while keeping
// the SigningKey intact. It will append the old identity to g.oldKeys.
func (g *Group) rekey(individualNotifications bool) error {
	// Create a new identity with a fresh EncryptionKey
	id, err := identity.CreateIdentity()
	if err != nil {
		return err
	}

	// Add the old identity to the list of oldKeys
	oldIdentity := g.identity
	g.oldKeys = append(g.oldKeys, oldIdentity)

	// Copy the old identity into a new identity.
	newID := &identity.Identity{}
	*newID = *g.identity

	// Move the new EncryptionKey into the new identity
	newID.EncryptionKey = id.EncryptionKey

	// Set the newID as _the_ identity for the group.
	g.identity = newID

	return g.rekeyNotifications(oldIdentity, individualNotifications)
}

func (g *Group) rekeyNotifications(old *identity.Identity, individualNotifications bool) error {
	// identify its messages.
	rekeyID := uuid.New()
	rekeyName := fmt.Sprintf("group-rekey/%s/%s", g.uniqueID, rekeyID)

	//
	var err error
	receivers := []*identity.Address{g.identity.Address}
	if individualNotifications {
		receivers, err = g.loadMembers()
		if err != nil {
			return err
		}
	}

	// Send Invitations
	invite := message.CreateMail(
		g.identity.Address,
		time.Now(),
		rekeyName,
		receivers...,
	)

	// Encode the Identity of the Group
	idBuffer := &bytes.Buffer{}
	_, err = g.identity.GobEncodeKey(idBuffer)
	if err != nil {
		return err
	}

	// Create the Rekeying Message
	invite.Components = message.ComponentList{
		groupInvite:        message.CreateStringComponent(groupInvite, g.name),
		groupRekeyIdentity: message.CreateComponent(groupRekeyIdentity, idBuffer.Bytes()),
		groupInviteMembers: message.CreateStringComponent(groupInviteMembers, strings.Join(g.members, ",")),
		groupInviteID:      message.CreateStringComponent(groupInviteID, g.uniqueID),
	}

	// Send out the invitation message.
	_, err = g.client().PublishMessage(
		invite,
		receivers,
		rekeyName,
		true,
	)
	return err
}

// Leave will remove the `user` from the Group g.
func (g *Group) Leave(user *identity.Identity, alert bool) error {
	return nil
}

// Kick will remove the `user` from the Group g.
func (g *Group) Kick(user string) error {
	// Remove the member from our list of members
	for i, v := range g.members {
		if v == user {
			g.members[i], g.members = g.members[len(g.members)-1], g.members[:len(g.members)-1]
			break
		}
	}

	// Rekey the group
	return g.rekey(true)
}

// Invite will invite a member to the Group g.
func (g *Group) Invite(user string, rekey bool) error {
	// Invite the user and add them to the list of members
	id, err := g.router.LookupAlias(user, routing.LookupTypeMAIL)
	if err != nil {
		return err
	}

	g.members = append(g.members, user)

	if err := g.inviteMembers(g.identity, id); err != nil {
		return err
	}

	if rekey {
		return g.rekey(false)
	}
	return nil
}

func RequestJoinGroup(address string) (*Group, error) {
	return nil, nil
}
