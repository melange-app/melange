package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"getmelange.com/zooko"
	"github.com/melange-app/nmcd/btcjson"
	"github.com/melange-app/nmcd/wire"
)

func main() {
	var (
		namecoinServer   = flag.String("namecoin", "", "The URL of the RPC Server.")
		namecoinUsername = flag.String("rpcusername", "", "The username to use when authenticating with the Namecoin RPC Server.")
		namecoinPassword = flag.String("rpcpassword", "", "The password to use when authenticating with the Namecoin RPC Server.")
	)

	// Go ahead and Parse the Flags
	flag.Parse()

	server := rpcServer{
		Username: *namecoinUsername,
		Password: *namecoinPassword,
		Host:     *namecoinServer,
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		// Get the command from stdin
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "quit" {
			return
		}

		fmt.Println("Looking up", text)

		nhistory, err := btcjson.NewNameHistoryCmd(nil, text)
		if err != nil {
			fmt.Println("Couldn't create Name History", err)
			continue
		}

		reply, err := server.Send(nhistory)
		if err != nil {
			fmt.Println("Couldn't send name_history.", err)
			continue
		}

		result := reply.Result.([]btcjson.NameInfoResult)
		fmt.Println("Found", len(result), "name entries.")

		// Go through each history and extract the transaction information.
		for _, entry := range result {
			// Get the Raw Transaction Associated with the Name Operation
			getRaw, err := btcjson.NewGetRawTransactionCmd(nil, entry.TX, 1)
			if err != nil {
				fmt.Println("Couldn't create transaction command for", entry.TX, err)
				continue
			}

			reply, err := server.Send(getRaw)
			if err != nil {
				fmt.Println("Got error getting transaction details", err)
				continue
			}

			transactionResult := reply.Result.(*btcjson.TxRawResult)
			fmt.Println("Found entry at block", transactionResult.BlockHash, transactionResult.Confirmations)

			// Get the Block Associated with the Name Operation
			getBlock, err := btcjson.NewGetBlockCmd(nil, transactionResult.BlockHash)
			if err != nil {
				fmt.Println("Couldn't create GetBlockCmd for", transactionResult.BlockHash, err)
				continue
			}

			reply, err = server.Send(getBlock)
			if err != nil {
				fmt.Println("got error getting block details", err)
				continue
			}

			block := reply.Result.(*btcjson.BlockResult)
			fmt.Println("Block has", len(block.Tx), "transactions.")

			// Create the Merkle Branch
			branch, err := constructMerkleTree(block.Tx, block.MerkleRoot, transactionResult.Txid)
			if err != nil {
				fmt.Println("Error constructing Merkle Tree", err)
				continue
			}

			root, _ := wire.NewShaHashFromStr(block.MerkleRoot)
			txid, _ := wire.NewShaHashFromStr(transactionResult.Txid)
			if err = zooko.VerifyMerkleBranch(branch, *root, *txid); err != nil {
				fmt.Println("The constructed Merkle Branch is invalid.", err)
			}
		}
	}
}

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
