package chain

const endpoint = "http://namecoin.webbtc.com"

type WebBTC struct{}

// LoadCredentials will load credentials into the software.
func (w *WebBTC) LoadCredentials(map[string]string) {}

// BroadcastTransaction will broadcast a transaction to the network.
func (w *WebBTC) BroadcastTransaction(tx []byte) bool {

}

// GetName will download the name data associated with a name.
func (w *WebBTC) GetName(name string) *Name {

}
