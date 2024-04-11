package test_util

import (
	"context"
	"encoding/base64"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	"github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/sgerogia/sol-stablecoin/tpp-client/encrypt"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"math/big"
	"os"
)

const INSTITUTION = "natwest-sandbox"
const AMOUNT = "1"
const NET_OF_SEIGNORAGE = 999
const ONE_HUND_PERC = 1000

func Payer() *bank.AccountDetails {
	return &bank.AccountDetails{
		Name:          "John Doe",
		SortCode:      "500000",
		AccountNumber: "12345601",
	}
}

func Receiver() *bank.AccountDetails {
	return &bank.AccountDetails{
		Name:          "ProvableGBP Limited",
		SortCode:      "500000",
		AccountNumber: "87654301",
	}
}

func MintRequestPayload(pubKey []byte) *event.MintRequestPayload {
	acc := Payer()
	return &event.MintRequestPayload{
		InstitutionId: INSTITUTION,
		AccountNumber: acc.AccountNumber,
		SortCode:      acc.SortCode,
		Name:          acc.Name,
		PublicKey:     base64.StdEncoding.EncodeToString(pubKey),
	}
}

type NatwestSandboxInfo struct {
	ClientId         string
	ClientSecret     string
	RedirectUrl      string
	CustomerUsername string
}

// GetNatwestSandboxInfo returns the information needed to connect to the Natwest sandbox or nil if the environment variables are not set
func GetNatwestSandboxInfo() *NatwestSandboxInfo {

	if !isEnvVarSet("NATWEST_SANDBOX_CLIENT_ID") ||
		!isEnvVarSet("NATWEST_SANDBOX_CLIENT_SECRET") ||
		!isEnvVarSet("NATWEST_SANDBOX_REDIRECT_URL") ||
		!isEnvVarSet("NATWEST_SANDBOX_CUSTOMER_USERNAME") {
		return nil
	}
	return &NatwestSandboxInfo{
		ClientId:         os.Getenv("NATWEST_SANDBOX_CLIENT_ID"),
		ClientSecret:     os.Getenv("NATWEST_SANDBOX_CLIENT_SECRET"),
		RedirectUrl:      os.Getenv("NATWEST_SANDBOX_REDIRECT_URL"),
		CustomerUsername: os.Getenv("NATWEST_SANDBOX_CUSTOMER_USERNAME"),
	}
}

func isEnvVarSet(e string) bool {
	return os.Getenv(e) != ""
}

type ChainInfo struct {
	TppContractClient   *contract.ContractClient
	TppKeyPair          *encrypt.KeyPair
	PayerContractClient *contract.ContractClient
	PayerKeyPair        *encrypt.KeyPair
	ContractAddress     *common.Address
	Backend             *backends.SimulatedBackend
}

func DeployProvableGBPAndCreateAccounts() (*ChainInfo, error) {

	// TPP key pair
	tppPair, err := encrypt.NewKeyPair()
	if err != nil {
		return nil, err
	}
	tppAuth, err := bind.NewKeyedTransactorWithChainID(tppPair.PrivateKey, big.NewInt(1337))
	if err != nil {
		return nil, err
	}

	// payer key pair
	payerPair, err := encrypt.NewKeyPair()
	if err != nil {
		return nil, err
	}
	payerAuth, err := bind.NewKeyedTransactorWithChainID(payerPair.PrivateKey, big.NewInt(1337))
	if err != nil {
		return nil, err
	}

	// genesis accounts
	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei

	tppAddress := tppAuth.From
	payerAddress := payerAuth.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		tppAddress: {
			Balance: balance,
		},
		payerAddress: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(4712388)
	simulClient := backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)

	// deploy contract and mint a block
	contractAddress, _, _, err := contract.DeployProvableGBP(tppAuth, simulClient, tppPair.PublicEncrKeyBytes()[:])
	if err != nil {
		return nil, err
	}
	simulClient.Commit()

	// create contract clients
	tppClient, err := NewTestContractClient(tppAuth, &contractAddress, simulClient)
	if err != nil {
		return nil, err
	}
	payerClient, err := NewTestContractClient(payerAuth, &contractAddress, simulClient)
	if err != nil {
		return nil, err
	}

	return &ChainInfo{
		TppContractClient:   tppClient,
		TppKeyPair:          tppPair,
		PayerContractClient: payerClient,
		PayerKeyPair:        payerPair,
		ContractAddress:     &contractAddress,
		Backend:             simulClient,
	}, nil
}

func NewTestContractClient(
	account *bind.TransactOpts,
	contractAddress *common.Address,
	client *backends.SimulatedBackend) (*contract.ContractClient, error) {

	contr, err := contract.NewProvableGBP(*contractAddress, client)
	if err != nil {
		return nil, err
	}

	filter, err := contract.NewProvableGBPFilterer(*contractAddress, client)
	if err != nil {
		return nil, err
	}

	return contract.NewContractClient2(
		client,
		client,
		&contract.ProvableGBPSession{
			Contract:     contr,
			TransactOpts: *account,
			CallOpts: bind.CallOpts{
				Pending: false,
				From:    (*account).From,
				Context: context.Background(),
			},
		},
		filter,
	), nil
}
