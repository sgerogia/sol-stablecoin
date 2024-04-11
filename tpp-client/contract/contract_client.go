package contract

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type ContractClient struct {
	contrTransactor bind.ContractTransactor
	trxReader       ethereum.TransactionReader
	session         *ProvableGBPSession
	events          *ProvableGBPFilterer
	contrAddress    common.Address
}

func NewContractClient(
	providerUrl string,
	chainId int64,
	contractAddress string,
	privateKey string,
	gasLimit int64) (*ContractClient, error) {

	client, err := ethclient.Dial(providerUrl)
	if err != nil {
		return nil, errors.New("Error connecting to Ethereum node: " + err.Error())
	}

	opts, fromAddr, err := KeysAndAddress(privateKey, chainId)
	if err != nil {
		return nil, err
	}

	opts.GasLimit = uint64(gasLimit)
	opts.Value = big.NewInt(0)

	contrAddr := common.HexToAddress(contractAddress)

	contr, err := NewProvableGBP(contrAddr, client)
	if err != nil {
		return nil, errors.New("Error creating contract caller instance: " + err.Error())
	}

	filter, err := NewProvableGBPFilterer(contrAddr, client)
	if err != nil {
		return nil, errors.New("Error creating contract event filter: " + err.Error())
	}

	return &ContractClient{
		contrTransactor: client,
		trxReader:       client,
		events:          filter,
		session: &ProvableGBPSession{
			Contract:     contr,
			TransactOpts: *opts,
			CallOpts: bind.CallOpts{
				Pending: false,
				From:    *fromAddr,
				Context: context.Background(),
			},
		},
	}, nil
}

// NewContractClient2 use only in testing.
func NewContractClient2(
	_ct bind.ContractTransactor,
	_tr ethereum.TransactionReader,
	_session *ProvableGBPSession,
	_events *ProvableGBPFilterer) *ContractClient {
	return &ContractClient{
		contrTransactor: _ct,
		trxReader:       _tr,
		session:         _session,
		events:          _events,
	}
}

// GetSingleUseSession returns a single use session, with a fresh nonce and gas price.
//
// WARNING: NON-PRODUCTION CODE
// This is useful for transactions that are not expected to be retried.
// If a transaction is retried, the nonce will be incremented and the transaction will fail.
// Similarly, if retried and the gas price has changed, the transaction will fail.
// Last but not least, in a highly parallel environment, the nonce may be incremented by another transaction.
// A much better approach would be to serialize (and monitor) outgoing transactions through a single "outbox".
func (_contrClient *ContractClient) GetSingleUseSession() (*ProvableGBPSession, error) {

	nonce, err := _contrClient.contrTransactor.PendingNonceAt(context.Background(), _contrClient.session.TransactOpts.From)
	if err != nil {
		return nil, errors.New("Error getting nonce: " + err.Error())
	}

	gasPrice, err := _contrClient.contrTransactor.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.New("Error getting gas price: " + err.Error())
	}

	return &ProvableGBPSession{
		Contract: _contrClient.session.Contract,
		CallOpts: _contrClient.session.CallOpts,
		TransactOpts: bind.TransactOpts{
			GasLimit: _contrClient.session.TransactOpts.GasLimit,
			Value:    _contrClient.session.TransactOpts.Value,
			Nonce:    big.NewInt(int64(nonce)),
			GasPrice: gasPrice,
			Signer:   _contrClient.session.TransactOpts.Signer,
			From:     _contrClient.session.TransactOpts.From,
			Context:  _contrClient.session.TransactOpts.Context,
		},
	}, nil
}

func (_contrClient *ContractClient) GetContractAddress() common.Address {
	return _contrClient.contrAddress
}

func (_contrClient *ContractClient) GetEventFilterer() *ProvableGBPFilterer {
	return _contrClient.events
}

// KeysAndAddress returns a contract signer (TransactOpts) and a public address, as derived from a private key
func KeysAndAddress(key string, chainId int64) (*bind.TransactOpts, *common.Address, error) {

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, nil, errors.New("Error converting private key to ECDSA: " + err.Error())
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		return nil, nil, errors.New("Error creating private key signer: " + err.Error())
	}

	return opts, &fromAddress, nil
}
