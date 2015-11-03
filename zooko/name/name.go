package name

// Status represents the different possible states that a Zooko
// registration could be in.
type Status int

const (
	// In this stage, the name has not yet been included and
	// revealed to the blockchain. This would require that the
	// Zooko server keep the registration in its memory in order
	// to resolve the name.

	// Broadcast represents the sending of the first name_new
	// command into the Namecoin network.
	Broadcast Status = iota

	// Confirmed represents the inclusion of the first name_new
	// command into the Namecoin blockchain.
	Confirmed

	// Once the name has been revealed, Zooko can look to the
	// blockchain to get the resolution. Until `Registered`
	// status, however, the name is not fully secure.

	// Revealed indicates that the name_firstupdate transaction
	// has been broadcast to the network. It may take some time
	// for this to be included in a block as it may have been sent
	// well before the 12 confirmations required to validate the
	// transaction.
	Revealed

	// Registered represents the final step in the name
	// registration saga. At this point, the name is fully
	// included in the Namecoin blockchain.
	Registered
)

// Registration gives a view to a single Zooko name registration.
type Registration struct {
	Name   string
	Data   []byte
	Status Status
}
