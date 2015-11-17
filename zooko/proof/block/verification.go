package proof

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/melange-app/nmcd/wire"
)

type verificationError struct {
	message   string
	errorType int
}

const (
	errUnexpectedDifficulty = iota
	errHighHash
	errInvalidMerkleTree
)

func (v *verificationError) Error() string {
	return v.message
}

func ruleError(errorType int, msg string) *verificationError {
	return &verificationError{
		message:   msg,
		errorType: errorType,
	}
}

func VerifyMerkleBranch(m wire.MerkleBranch, root *wire.ShaHash, check *wire.ShaHash) (*wire.ShaHash, error) {
	var working *wire.ShaHash = new(wire.ShaHash)
	*working = *check
	for i, v := range m.BranchHash {
		side := m.BranchSideMask & (1 << uint(i))

		if side == 0 {
			// Zero means it goes on the right
			_ = working.SetBytes(wire.DoubleSha256(append(working[0:], v[0:]...)))
		} else {
			// One means it goes on the left
			_ = working.SetBytes(wire.DoubleSha256(append(v[0:], working[0:]...)))
		}
	}

	if root == nil {
		return working, nil
	}

	if working.IsEqual(root) {
		return nil, nil
	}

	return nil, errors.New("Merkle tree is not valid.")
}

func verifyProofOfWork(hdr *wire.BlockHeader, claimedTarget uint32, powLimit *big.Int) error {
	// The target difficulty must be larger than zero.
	target := compactToBig(claimedTarget)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too low",
			target)
		return ruleError(errUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(powLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, powLimit)
		return ruleError(errUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target.
	blockHash, err := hdr.BlockSha()
	if err != nil {
		return err
	}
	hashNum := shaHashToBig(&blockHash)
	if hashNum.Cmp(target) > 0 {
		str := fmt.Sprintf("block hash of %064x is higher than "+
			"expected target of %064x", hashNum, target)
		return ruleError(errHighHash, str)
	}

	return nil
}

// 0x039ea604040001e80359124d696e656420627920425443204775696c642c
// fabe6d6d {0xfabe, "m", "m"}
// e66de6467b84e3647bdaf9b1ecf0c6cc2f321798abe704e5666a8cac96903e5e
// 01000000
// 00000000
// 08000000820000000e

const blockVersionChainStart = (1 << 16)

func verifiyBlockHeader(previous *wire.BlockHeader, current *wire.BlockHeader) error {
	if previous != nil {
		previousSha, err := previous.BlockSha()
		if err != nil {
			return err
		}

		// Check the the current block and the previous block are linked
		if !previousSha.IsEqual(&current.PrevBlock) {
			return errors.New("Block has incorrect previous block sha.")
		}
	}

	if current.AuxPowHeader == nil {
		// If there is no AuxPowHeader, then we need to validate the
		// Header's Hash traditionally
		return verifyProofOfWork(
			current,
			current.Bits,
			mainPowLimit,
		)
	} else {
		// Otherwise, we need to validate the AuxPowHeader...
		//
		err := verifyProofOfWork(
			&current.AuxPowHeader.ParentBlock,
			current.Bits,
			mainPowLimit,
		)
		if err != nil {
			return err
		}

		// Verify that the CoinbaseTx is in the Merkle Tree...
		coinbaseSha, err := current.AuxPowHeader.CoinbaseTx.TxSha()
		if err != nil {
			return err
		}

		_, err = VerifyMerkleBranch(
			current.AuxPowHeader.CoinbaseBranch,
			&current.AuxPowHeader.ParentBlock.MerkleRoot,
			&coinbaseSha,
		)
		if err != nil {
			return err
		}

		if len(current.AuxPowHeader.CoinbaseTx.TxIn) < 1 {
			return errors.New("No coinbase transaction...")
		}

		currentSha, err := current.BlockSha()
		if err != nil {
			return err
		}

		cbRoot, _ := VerifyMerkleBranch(
			current.AuxPowHeader.BlockchainBranch,
			nil,
			&currentSha,
		)

		coinbase := current.AuxPowHeader.CoinbaseTx.TxIn[0]
		mm, err := readMergedMiningTransaction(coinbase, &currentSha, cbRoot)
		if err != nil {
			return err
		}

		if mm.MerkleNonce == 0 && mm.MerkleSize == 1 {
			if !currentSha.IsEqual(&mm.BlockHash) {
				return fmt.Errorf(
					"Coinbase TX's hash (%s) doesn't match block's (%s).",
					mm.BlockHash,
					currentSha,
				)
			}
		} else {
			if _, err := VerifyMerkleBranch(
				current.AuxPowHeader.BlockchainBranch,
				&mm.BlockHash,
				&currentSha,
			); err != nil {
				return err
			}
		}

	}

	return nil
}

type mergedMiningTransaction struct {
	// BlockHash is the hash of the AuxPow Block Header
	BlockHash wire.ShaHash

	// MerkleSize is the number of entries in the aux work
	// merkle tree (probably just 1).
	MerkleSize int32

	// MerkleNonce is the nonce use to calculate indexes.
	// Generally left as 0.
	MerkleNonce int32
}

var mergedMiningMagicBytes = []byte{0xfa, 0xbe, 'm', 'm'}

func readMergedMiningTransaction(t *wire.TxIn, blockHash *wire.ShaHash, rootMerkleHash *wire.ShaHash) (*mergedMiningTransaction, error) {
	script := t.SignatureScript

	// Look for the magic number in the SignatureScript
	idx := bytes.Index(script, mergedMiningMagicBytes)

	if idx == -1 {
		blockBytes, _ := hex.DecodeString(blockHash.String())
		idx = bytes.Index(script, blockBytes)

		if idx == -1 {
			blockBytes, _ := hex.DecodeString(rootMerkleHash.String())
			idx = bytes.Index(script, blockBytes)

			if idx == -1 {
				return nil, errors.New("Couldn't find the magic bytes in the signature script.")
			}

		}

		idx = idx - len(mergedMiningMagicBytes)
	}

	data := script[idx+len(mergedMiningMagicBytes):]

	blockHashData := hex.EncodeToString(data[:32])
	blockHash, err := wire.NewShaHashFromStr(blockHashData)
	if err != nil {
		return nil, err
	}

	merkleSizeData := data[32:36]
	merkleNonceData := data[36:40]
	return &mergedMiningTransaction{
		BlockHash:   *blockHash,
		MerkleSize:  int32(binary.LittleEndian.Uint32(merkleSizeData)),
		MerkleNonce: int32(binary.LittleEndian.Uint32(merkleNonceData)),
	}, nil
}
