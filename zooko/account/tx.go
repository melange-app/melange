package account

import (
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/melange-app/nmcd/btcutil"
	"github.com/melange-app/nmcd/chaincfg"
	"github.com/melange-app/nmcd/txscript"
	"github.com/melange-app/nmcd/wire"
)

const (
	// TxFee in Namecoin is set to be 5mNMC (or 5 * 10^5
	// nmc-satoshis).
	txFee                     = 5e5
	defaultTransactionVersion = 1
)

// Transaction object holds a transaction as we are building it. It
// doesn't update any of the UTXO until it is actually broadcast to
// the network.
type Transaction struct {
	// Transaction contains a bitcoin wire Transaction
	*wire.MsgTx

	// We also keep track of which UTXO objects we are spending
	// and creating.
	Spent []*UTXO
	New   []*UTXO

	// It includes a total amount.
	Amount int64
}

// TransferFunds will create a transaction that transfers an amount of
// NMC-satoshis to a hashed public key.
func (a *Account) TransferFunds(amount int64, pubkeyHash string) (*Transaction, error) {
	// Decode the hex representation of the Public Key.
	hashBytes, err := hex.DecodeString(pubkeyHash)
	if err != nil {
		return nil, err
	}

	// Convert it to a btcd address object.
	addr, err := btcutil.NewAddressPubKeyHash(hashBytes, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	// Generate the pkScript for this output.
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

	// That is the only output here.
	txOutput := []*wire.TxOut{
		wire.NewTxOut(amount, pkScript),
	}

	// Build the transaction.
	return a.buildTransaction(txOutput)
}

// buildTransaction will attempt to create a new transaction using the
// UTXO that the account has stored.
func (a *Account) buildTransaction(output []*wire.TxOut) (*Transaction, error) {
	// Use the default transaction version = 1.
	return a.buildTransactionVersion(output, defaultTransactionVersion)
}

func (a *Account) buildTransactionVersion(output []*wire.TxOut, version int32) (*Transaction, error) {
	// Calculate the amount of the transaction.
	var amount int64
	for _, v := range output {
		amount += v.Value
	}

	// Add the transaction fee to the outputs - we will not
	// redirect this amount in the change output.
	amount += txFee

	// Ensure that we have enough money in the account to make the transaction.
	if a.Balance() < amount {
		return nil, fmt.Errorf("zooko: balance (%d) is too low to make transaction (%d)",
			a.Balance(), amount)
	}

	// Sort the transactions list by amount so that we pick the
	// highest.
	sort.Sort(a.Unspent)

	msgTx := wire.NewMsgTx()
	msgTx.Version = version

	var balance int64
	var toSpend UTXOList
	for balance < amount {
		// Get the next highest transaction
		newTx := a.Unspent[len(toSpend)]

		hash, err := wire.NewShaHashFromStr(newTx.TxID)
		if err != nil {
			return nil, err
		}

		// Convert the UTXO to a TxIn and add to the MsgTx. Currently no scriptSig.
		msgTx.AddTxIn(
			wire.NewTxIn(wire.NewOutPoint(hash, newTx.Output), nil),
		)

		// Update our structures
		toSpend = append(toSpend, newTx)
		balance += newTx.Amount
	}

	msgTx.TxOut = output

	// Build a change transaction.
	hasChange := false
	if amount < balance {
		change := balance - amount
		hasChange = true

		addr, err := a.PublicKeyHash()
		if err != nil {
			return nil, err
		}

		// Create the pay to public key hash script.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}

		// Add the change transaction.
		msgTx.AddTxOut(wire.NewTxOut(change, pkScript))
	}

	// We will now create the SigScript for each of the TxIn.
	for index, input := range toSpend {
		sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams,
			msgTx, index, input.PkScript, txscript.SigHashAll,
			a, nil, nil)
		if err != nil {
			return nil, err
		}

		// Set the signature script on the transaction.
		msgTx.TxIn[index].SignatureScript = sigScript
	}

	// Provide a list of new unspent transactions if we happened
	// to include a change transaction.
	var newUnspent []*UTXO
	if hasChange {
		// The change transaction is always the last index.
		changeIndex := len(msgTx.TxOut) - 1
		change := msgTx.TxOut[changeIndex]

		dataId, err := msgTx.TxSha()
		if err != nil {
			return nil, err
		}

		// Create the new Unspent Transaction Output
		newUnspent = append(newUnspent, &UTXO{
			TxID:     dataId.String(),
			Output:   uint32(changeIndex),
			Amount:   change.Value,
			PkScript: change.PkScript,
		})
	}

	// We must create a change transaction as a TxOut to ourselves
	// and save that as a new UTXO.
	return &Transaction{
		MsgTx: msgTx,

		Spent: toSpend,
		New:   newUnspent,

		Amount: balance,
	}, nil
}
