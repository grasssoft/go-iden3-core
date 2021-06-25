package idenpubonchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	zktypes "github.com/iden3/go-circom-prover-verifier/types"
	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/core/proof"
	"github.com/iden3/go-iden3-core/eth"
	"github.com/iden3/go-iden3-core/eth/contracts"
	zkutils "github.com/iden3/go-iden3-core/utils/zk"
	"github.com/iden3/go-merkletree"
)

var (
	ErrIdenNotOnChain              = fmt.Errorf("Identity not found on chain")
	ErrIdenNotOnChainOrBlockTooNew = fmt.Errorf("Identity not found on chain or the queried block number is not yet on chain")
	ErrIdenNotOnChainOrTimeTooNew  = fmt.Errorf("Identity not found on chain or the queried time is not yet on chain")
	ErrIdenByBlockNotFound         = fmt.Errorf("Identity not found by the queried block number")
	ErrIdenByTimeNotFound          = fmt.Errorf("Identity not found by the queried block timestamp")
)

// IdenPubOnChainer is an interface that gives access to the IdenStates Smart Contract.
type IdenPubOnChainer interface {
	GetState(id *core.ID) (*proof.IdenStateData, error)
	GetStateByBlock(id *core.ID, blockN uint64) (*proof.IdenStateData, error)
	GetStateByTime(id *core.ID, blockTimestamp int64) (*proof.IdenStateData, error)
	SetState(id *core.ID, newState *merkletree.Hash, proof *zktypes.Proof) (*types.Transaction, error)
	InitState(id *core.ID, genesisState *merkletree.Hash,
		newState *merkletree.Hash, proof *zktypes.Proof) (*types.Transaction, error)
	TxConfirmBlocks(tx *types.Transaction) (*big.Int, error)
	// VerifyProofClaim(pc *proof.ProofClaim) (bool, error)
}

// ContractAddresses are the list of Smart Contract addresses used for the on chain identity state data.
type ContractAddresses struct {
	IdenStates common.Address
}

// IdenPubOnChain is the regular implementation of IdenPubOnChain
type IdenPubOnChain struct {
	client    *eth.Client
	addresses ContractAddresses
}

// New creates a new IdenPubOnChain
func New(client *eth.Client, addresses ContractAddresses) *IdenPubOnChain {
	return &IdenPubOnChain{
		client:    client,
		addresses: addresses,
	}
}

// GetState returns the Identity State Data of the given ID from the IdenStates Smart Contract.
func (ip *IdenPubOnChain) GetState(id *core.ID) (*proof.IdenStateData, error) {
	var _idenState *big.Int
	var blockN uint64
	var blockTS uint64
	if err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, _idenState, err = idenStates.GetStateDataById(nil, id.BigInt())
		return err
	}); err != nil {
		return nil, err
	}
	idenState := merkletree.NewHashFromBigInt(_idenState)
	if idenState.Equals(&merkletree.HashZero) {
		return nil, ErrIdenNotOnChain
	}
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: idenState,
	}, nil
}

// GetStateByBlock returns the Identity State Data of the given ID published at
// queryBlockN from the IdenStates Smart Contract.
func (ip *IdenPubOnChain) GetStateByBlock(id *core.ID, queryBlockN uint64) (*proof.IdenStateData, error) {
	idenStateData, err := ip.GetStateClosestToBlock(id, queryBlockN)
	if err != nil {
		return nil, err
	}
	if idenStateData.BlockN != queryBlockN {
		return nil, ErrIdenByBlockNotFound
	}
	return idenStateData, nil
}

// GetStateClosestToBlock returns the Identity State Data of the given ID that
// is closest (equal or older) to the queryBlockN from the IdenStates Smart
// Contract.  If a resut is found, BlockN <= queryBlockN.
func (ip *IdenPubOnChain) GetStateClosestToBlock(id *core.ID, queryBlockN uint64) (*proof.IdenStateData, error) {
	var _idenState *big.Int
	var blockN uint64
	var blockTS uint64
	if err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, _idenState, err = idenStates.GetStateDataByBlock(nil,
			id.BigInt(), queryBlockN)
		return err
	}); err != nil {
		return nil, err
	}
	idenState := merkletree.NewHashFromBigInt(_idenState)
	if idenState.Equals(&merkletree.HashZero) {
		return nil, ErrIdenNotOnChainOrBlockTooNew
	}
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: idenState,
	}, nil
}

// GetStateByTime returns the Identity State Data of the given ID published at
// queryBlockTs from the IdenStates Smart Contract.
func (ip *IdenPubOnChain) GetStateByTime(id *core.ID, queryBlockTs int64) (*proof.IdenStateData, error) {
	idenStateData, err := ip.GetStateClosestToTime(id, queryBlockTs)
	if err != nil {
		return nil, err
	}
	if idenStateData.BlockTs != queryBlockTs {
		return nil, ErrIdenByTimeNotFound
	}
	return idenStateData, nil
}

// GetStateClosestToTime returns the Identity State Data of the given ID
// closest (equal or older) to the queryBlockTs from the IdenStates Smart
// Contract.  If a resut is found, BlockN <= queryBlockN.
func (ip *IdenPubOnChain) GetStateClosestToTime(id *core.ID, queryBlockTs int64) (*proof.IdenStateData, error) {
	var _idenState *big.Int
	var blockN uint64
	var blockTS uint64
	if err := ip.client.Call(func(c *ethclient.Client) error {
		idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
		if err != nil {
			return err
		}
		blockN, blockTS, _idenState, err = idenStates.GetStateDataByTime(nil,
			id.BigInt(), uint64(queryBlockTs))
		return err
	}); err != nil {
		return nil, err
	}
	idenState := merkletree.NewHashFromBigInt(_idenState)
	if idenState.Equals(&merkletree.HashZero) {
		return nil, ErrIdenNotOnChainOrTimeTooNew
	}
	return &proof.IdenStateData{
		BlockN:    blockN,
		BlockTs:   int64(blockTS),
		IdenState: idenState,
	}, nil
}

// InitState initializes the first Identity State of the given ID in the IdenStates Smart Contract.
func (ip *IdenPubOnChain) InitState(id *core.ID, genesisState *merkletree.Hash,
	newState *merkletree.Hash, proof *zktypes.Proof) (*types.Transaction, error) {
	if tx, err := ip.client.CallAuth(
		1000000,
		func(c *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
			if err != nil {
				return nil, err
			}
			proofA, proofB, proofC := zkutils.ProofToBigInts(proof)
			return idenStates.InitState(auth, newState.BigInt(),
				genesisState.BigInt(), id.BigInt(),
				proofA, proofB, proofC)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed initalizating identity state in the Smart Contract (initState): %w", err)
	} else {
		return tx, nil
	}
}

// SetState updates the Identity State of the given ID in the IdenStates Smart Contract.
func (ip *IdenPubOnChain) SetState(id *core.ID, newState *merkletree.Hash,
	proof *zktypes.Proof) (*types.Transaction, error) {
	if tx, err := ip.client.CallAuth(
		1000000,
		func(c *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			idenStates, err := contracts.NewState(ip.addresses.IdenStates, c)
			if err != nil {
				return nil, err
			}
			proofA, proofB, proofC := zkutils.ProofToBigInts(proof)
			return idenStates.SetState(auth, newState.BigInt(), id.BigInt(),
				proofA, proofB, proofC)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting identity state in the Smart Contract (setState): %w", err)
	} else {
		return tx, nil
	}
}

// TxConfirmBlocks returns the number of confirmed blocks of transaction tx.
func (ip *IdenPubOnChain) TxConfirmBlocks(tx *types.Transaction) (*big.Int, error) {
	receipt, err := ip.client.GetReceipt(tx)
	if err != nil {
		return nil, err
	}
	currentBlock, err := ip.client.CurrentBlock()
	if err != nil {
		return nil, err
	}
	return currentBlock.Sub(currentBlock, receipt.BlockNumber), nil
}
