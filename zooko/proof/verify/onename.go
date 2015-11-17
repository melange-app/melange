package chain

const endpoint = "https://api.onename.com/v1/"

type OneName struct {
	AppId     string
	AppSecret string
}

// BroadcastTransaction will take a signed Namecoin transaction and
// send it to the network.
func (o *OneName) BroadcastTransaction(tx []byte) bool {

}

// GetName would normally return the name data associated with a name,
// but OneName does not allow for arbitrary lookups of Namecoin data.
func (o *OneName) GetName(name string) *Name {
	return nil
}
