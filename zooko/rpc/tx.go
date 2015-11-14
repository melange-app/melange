package rpc

import (
	"encoding/hex"

	"getmelange.com/zooko/account"
	"github.com/melange-app/nmcd/btcjson"
)

// Endow will transfer money for a single Namecoin name registration
// (for 12 months) to a specified address. It will return the
// transaction details so that light clients can keep track of their
// UTXO.
func (r *Server) Endow(address string) (*account.UTXO, error) {
	// Send money to the address to endow
	sendCmd, err := btcjson.NewSendToAddressCmd(nil, address, endowmentAmount)
	if err != nil {
		return nil, err
	}

	reply, err := r.Send(sendCmd)
	if err != nil {
		return nil, err
	}

	if reply.Result == nil {
		return nil, errNilReply
	} else if reply.Error != nil {
		// Error from the namecoin daemon. This generally
		// indicates an internal server issue.
		return nil, *reply.Error
	}

	txid := reply.Result.(string)

	// Get information about the transaction that we just sent
	rawTxCmd, err := btcjson.NewGetRawTransactionCmd(nil, txid, 1)
	if err != nil {
		return nil, err
	}

	reply, err = r.Send(rawTxCmd)
	if err != nil {
		return nil, err
	}

	if reply.Result == nil {
		return nil, errNilReply
	} else if reply.Error != nil {
		// Error from the namecoin daemon. This generally
		// indicates an internal server issue.
		return nil, *reply.Error
	}

	rawTx := reply.Result.(*btcjson.TxRawResult)

	outputId := 0
	pkScript := ""
vOutLoop:
	for index, out := range rawTx.Vout {
		for _, addr := range out.ScriptPubKey.Addresses {
			if addr == address {
				outputId = index
				pkScript = out.ScriptPubKey.Hex
				break vOutLoop
			}
		}
	}

	pkData, err := hex.DecodeString(pkScript)
	if err != nil {
		return nil, err
	}

	return &account.UTXO{
		TxID:     txid,
		Output:   uint32(outputId),
		Amount:   endowmentAmount,
		PkScript: pkData,
	}, nil
}

// Broadcast will broadcast a raw hex-encoded transaction to the
// Namecoin network.
func (r *Server) Broadcast(tx string) error {
	cmd, err := btcjson.NewSendRawTransactionCmd(nil, tx)
	if err != nil {
		return err
	}

	result, err := r.Send(cmd)
	if result.Error != nil {
		return *result.Error
	}

	return err
}

func (r *Server) Confirmations(tx string) (int, error) {
	cmd, err := btcjson.NewGetTransactionCmd(nil, tx)
	if err != nil {
		return -1, err
	}

	result, err := r.Send(cmd)
	if result.Error != nil {
		return -1, *result.Error
	}

	txResult := result.Result.(*btcjson.GetTransactionResult)

	return int(txResult.Confirmations), nil
}
