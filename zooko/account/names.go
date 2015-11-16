package account

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/melange-app/nmcd/btcutil"
	"github.com/melange-app/nmcd/txscript"
	"github.com/melange-app/nmcd/wire"
)

const (
	// Name network fee is currently 10 mNMC.
	nameNetworkFee         = 10e5
	nameTransactionVersion = 28928
)

// ... << OP_DUP << OP_HASH160 << hash160 << OP_EQUALVERIFY << OP_CHECKSIG
func (a *Account) finishNameTransaction(name string, sb *txscript.ScriptBuilder, txIn []*UTXO) (*Transaction, error) {
	pk := a.Keys.Key().PubKey().SerializeCompressed()

	// Add the normal pk2pkh script at the end.
	sb.AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(btcutil.Hash160(pk)).
		AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG)

	script, err := sb.Script()
	if err != nil {
		return nil, err
	}
	txOutput := []*wire.TxOut{
		wire.NewTxOut(nameNetworkFee, script),
	}

	// Build the transaction
	finalTx, err := a.buildTransactionVersion(txOutput, txIn, nameTransactionVersion)
	if err != nil {
		return nil, err
	}

	hash, err := finalTx.MsgTx.TxSha()
	if err != nil {
		return nil, err
	}

	// Give the finalTx information about its name status for
	// commitment purposes.
	finalTx.Name = name
	finalTx.NameOut = &UTXO{
		TxID:     hash.String(),
		Output:   0,
		Amount:   nameNetworkFee,
		PkScript: script,
	}

	return finalTx, nil
}

// CreateNameNew will make a new transaction that preregisters a
// namecoin name in the blockchain.
//
// OP_NAME_NEW (1) << hash << OP_2DROP << ...
func (a *Account) CreateNameNew(name string) (*Transaction, []byte, error) {
	sb := txscript.NewScriptBuilder()

	// read a 64-bit random number to use as the salt
	buffer := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buffer)
	if err != nil {
		return nil, nil, err
	}

	// Hash the name with the salt.
	nameBy := []byte(name)
	hash := btcutil.Hash160(append(buffer, nameBy...))

	// Build the name script
	sb.AddOp(txscript.OP_1).AddData(hash).AddOp(txscript.OP_2DROP)

	tx, err := a.finishNameTransaction(name, sb, nil)
	if err != nil {
		return nil, nil, err
	}

	return tx, buffer, err
}

// CreateNameFirstUpdate will reveal a name and initial value to the
// blockchain.
//
// OP_NAME_FIRSTUPDATE (2) << vchName << vchRand << vchValue << OP_2DROP << OP_2DROP << ...
func (a *Account) CreateNameFirstUpdate(rand []byte, name, value string) (*Transaction, error) {
	sb := txscript.NewScriptBuilder()

	// Build the name script
	sb.AddOp(txscript.OP_2).
		AddData([]byte(name)).AddData(rand).AddData([]byte(value)).
		AddOp(txscript.OP_2DROP).AddOp(txscript.OP_2DROP)

	lastTx, err := a.getLastNameTx(name)
	if err != nil {
		return nil, err
	}

	return a.finishNameTransaction(name, sb, lastTx)
}

// CreateNameUpdate will make a new transation that updates a name to
// a new value.
//
// OP_NAME_UPDATE (3) << vchName << vchValue << OP_2DROP << OP_DROP << ...
func (a *Account) CreateNameUpdate(name, value string) (*Transaction, error) {
	sb := txscript.NewScriptBuilder()

	// Build the name script
	sb.AddOp(txscript.OP_3).
		AddData([]byte(name)).AddData([]byte(value)).
		AddOp(txscript.OP_2DROP).AddOp(txscript.OP_DROP)

	lastTx, err := a.getLastNameTx(name)
	if err != nil {
		return nil, err
	}

	return a.finishNameTransaction(name, sb, lastTx)
}

func (a *Account) getLastNameTx(name string) ([]*UTXO, error) {
	txList, ok := a.NameTx[name]
	if !ok || len(txList) < 1 {
		return nil, errors.New("zooko: cannot update a name without previously registering that name")
	}

	return []*UTXO{
		txList[len(txList)-1],
	}, nil
}
