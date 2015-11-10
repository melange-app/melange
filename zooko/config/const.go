package config

import (
	"encoding/hex"
	"fmt"

	"airdispat.ch/identity"
)

const (
	// These constants define the Zooko server powering Melange
	// and allow for secure communication. Currently, the Zooko
	// responses are not proof-carrying, but they will be in time.
	ServerLocation = "mailserver.airdispatch.me:4763"
	serverKey      = "4eff810301010e656e636f6465644164647265737301ff82000104010a456e637279" +
		"7074696f6e010a0001075369676e696e67010a0001084c6f636174696f6e010c000105416c696" +
		"173010c000000fe0160ff8201fe011641442d52534100000008000000000001000100000100d1" +
		"73a79c3587b86cacf06317048d90546bfd158466919f396fdb9742515d0840f64790730163445" +
		"aa54f201b7919fbfb892e2b20b4bfc57d47bc66d2684e045111eebe333ff36843ba3773216a06" +
		"80edbcebca07e38e82a4d0036b4f2179723016869543b8192e64a37220241267d2850cae62363" +
		"de2d223f8f918ba3a08027219b373babb01ea2bd20ffa029f5cd04efcab8c928fc8d3fa9417c1" +
		"a8c7144a04c4f90cbcc9af310e71d4612bd09b57a8b157c1e8e1c9b5157e524b4ccf913e85e37" +
		"92dfcee6d0c07a40460b38efab4f378a3de21fd6a66d29d9d7b495e334d11d78a5955812855f9" +
		"9b5aa65fe7ad0f5a86e53adc4f723862058f639467688b2f0141034d2f28a377c63bf45df1385" +
		"32795d5cb692f235b1886f4dfd01f6fc114f15eb5dcaeab17475f95052b981f081a42423ad7a1" +
		"9dfcdaafe6d60efc43bccbb9a57800"
)

var cachedAddress *identity.Address

// GetServerAddress will return the AirDispatch
func ServerAddress() *identity.Address {
	if cachedAddress == nil {
		data, err := hex.DecodeString(serverKey)
		if err != nil {
			fmt.Println("Error decoding data", err)
		}

		cachedAddress, err = identity.DecodeAddress(data)
		if err != nil {
			fmt.Println("Error decoding address", err)
		}

		cachedAddress.Location = ServerLocation
	}

	return cachedAddress
}
