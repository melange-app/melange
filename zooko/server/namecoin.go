package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"getmelange.com/zooko"
	"getmelange.com/zooko/message"
	"github.com/melange-app/nmcd/btcjson"
	"github.com/melange-app/nmcd/wire"
)

func (server *ZookoServer) CreateTransactionListForName(name string, excludeExpired bool) ([]message.NamecoinTransaction, error) {
	nhistory, err := btcjson.NewNameHistoryCmd(nil, name)
	if err != nil {
		return nil, err
	}

	reply, err := server.Send(nhistory)
	if err != nil {
		return nil, err
	}

	result := reply.Result.([]btcjson.NameInfoResult)

	// Go through each history and extract the transaction information.
	var transactions []message.NamecoinTransaction
	for _, entry := range result {
		// Get the Raw Transaction Associated with the Name Operation
		getRaw, err := btcjson.NewGetRawTransactionCmd(nil, entry.TX, 1)
		if err != nil {
			return nil, err
		}

		reply, err := server.Send(getRaw)
		if err != nil {
			return nil, err
		}

		transactionResult := reply.Result.(*btcjson.TxRawResult)
		// fmt.Println("Found entry at block", transactionResult.BlockHash, transactionResult.Confirmations)

		// Get the Block Associated with the Name Operation
		getBlock, err := btcjson.NewGetBlockCmd(nil, transactionResult.BlockHash)
		if err != nil {
			return nil, err
		}

		reply, err = server.Send(getBlock)
		if err != nil {
			return nil, err
		}

		block := reply.Result.(*btcjson.BlockResult)
		fmt.Println("Block has", len(block.Tx), "transactions.")

		// Create the Merkle Branch
		branch, err := constructMerkleTree(block.Tx, block.MerkleRoot, transactionResult.Txid)
		if err != nil {
			return nil, err
		}

		root, _ := wire.NewShaHashFromStr(block.MerkleRoot)
		txid, _ := wire.NewShaHashFromStr(transactionResult.Txid)
		if _, err = zooko.VerifyMerkleBranch(branch, root, txid); err != nil {
			return nil, errors.New("unable to verify new merkle branch")
		}

		bytesTxResult, err := json.Marshal(transactionResult)
		if err != nil {
			return nil, err
		}

		var outHashes [][]byte
		for _, v := range branch.BranchHash {
			outHashes = append(outHashes, v.Bytes())
		}

		transactions = append(transactions, message.NamecoinTransaction{
			TxId:               transactionResult.Txid,
			Branch:             branch.BranchSideMask,
			VerificationHashes: outHashes,
			Raw:                bytesTxResult,
			BlockId:            transactionResult.BlockHash,
		})
	}

	return transactions, nil
}

// constructMerkleTree will create a wire.MerkleBranch object that places the
// transaction in the block Merkle Tree.
func constructMerkleTree(transactions []string, root string, transaction string) (wire.MerkleBranch, error) {
	workingRow := make([]*wire.ShaHash, len(transactions))

	// Used to build the Merkle Branch
	var (
		currentTransaction int
		verificationHashes []wire.ShaHash
		height             uint32
		merkleBitmask      int32
	)

	var err error

	for i, v := range transactions {
		if v == transaction {
			currentTransaction = i
		}

		if workingRow[i], err = wire.NewShaHashFromStr(v); err != nil {
			return wire.MerkleBranch{}, err
		}
	}

	for {
		// We double the last entry if the row has an odd number of
		if len(workingRow)%2 == 1 {
			workingRow = append(workingRow, workingRow[len(workingRow)-1])
		}

		// Hash into the Next Row
		var nextRow []*wire.ShaHash
		for i := 0; i < len(workingRow); i += 2 {
			left := workingRow[i]
			right := workingRow[i+1]

			if currentTransaction == i {
				// We need to add the right hash to the merkle branch.
				verificationHashes = append(verificationHashes, *right)
				currentTransaction = len(nextRow)
				merkleBitmask = merkleBitmask | (0 << height)
			} else if currentTransaction == i+1 {
				// We need to add the left hash to the merkle branch.
				verificationHashes = append(verificationHashes, *left)
				currentTransaction = len(nextRow)
				merkleBitmask = merkleBitmask | (1 << height)
			}

			hash, err := wire.NewShaHash(wire.DoubleSha256(
				append(left.Bytes(), right.Bytes()...),
			))
			if err != nil {
				return wire.MerkleBranch{}, err
			}

			nextRow = append(nextRow, hash)
		}

		if len(nextRow) == 1 {
			if nextRow[0].String() != root {
				return wire.MerkleBranch{}, errors.New("Cannot validate MerkleTree.")
			}

			break
		}

		workingRow = nextRow
		height++
	}

	return wire.MerkleBranch{
		BranchHash:     verificationHashes,
		BranchSideMask: merkleBitmask,
	}, nil
}
