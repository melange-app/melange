package resolver

import (
	"errors"

	"code.google.com/p/goprotobuf/proto"

	adErrors "airdispat.ch/errors"
	adMessage "airdispat.ch/message"
	"airdispat.ch/wire"

	"getmelange.com/zooko/account"
	"getmelange.com/zooko/config"
	"getmelange.com/zooko/message"
)

func (c *Client) checkAccountBalance(acc *account.Account, registration bool) error {
	minimum := int64(account.NameMinimumBalance)
	if registration {
		// If we are doing the first registration of a name,
		// we must have enough coins for two of the
		// transactions.
		minimum *= 2
	}

	if acc.Balance() < minimum {
		// We must endow the account with funds.

		// Grab the namecoin address of the account
		hash, err := acc.PublicKeyHash()
		if err != nil {
			return err
		}
		addr := hash.String()

		endowRequest, err := message.CreateMessage(&message.RequestFunds{
			Address: &addr,
		}, c.Origin, config.ServerAddress())
		if err != nil {
			return err
		}

		data, typ, h, err := adMessage.SendMessageAndReceiveWithTimestamp(
			endowRequest,
			c.Origin, config.ServerAddress())
		if err != nil {
			return err
		} else if typ == wire.ErrorCode {
			return adErrors.CreateErrorFromBytes(data, h)
		} else if typ != message.TypeTransferFunds {
			return errors.New("zooko/resolver: endowment received incorrect response type")
		}

		transferredFunds := new(message.TransferFunds)
		if err := proto.Unmarshal(data, transferredFunds); err != nil {
			return err
		}

		// Build the UTXO for the account
		acc.Unspent = append(acc.Unspent, &account.UTXO{
			TxID:     *transferredFunds.Id,
			Output:   *transferredFunds.Index,
			PkScript: transferredFunds.Script,
			Amount:   *transferredFunds.Amount,
		})

		// Go ahead and ensure that the account is now over the minimum
		return c.checkAccountBalance(acc, registration)
	}

	// Otherwise, we are good to go.
	return nil
}
