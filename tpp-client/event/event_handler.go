package event

import (
	"encoding/hex"
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	"github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/shopspring/decimal"
	"math/big"
)

const DECIMAL_DIGITS = 18

// EventHandler are methods processing inbound events (contract and bank responses)
type EventHandler interface {

	// ProcessMintRequest processes a `MintRequest` event coming from the payer via the contract.
	// It will create a pending Consent request at the bank and call the contract's `AuthRequest` method
	ProcessMintRequest(request *contract.ProvableGBPMintRequest) error

	// ProcessAuthGranted processes an AuthGranted event coming from the payer via the contract.
	// I.e. the payer has authorised the payment.
	// The handler will submit the payment to the bank and schedule the check of the payment's final settlement.
	ProcessAuthGranted(request *contract.ProvableGBPAuthGranted) error

	// ProcessPaymentStatusResponse called by the scheduler when a payment status response is received from the bank.
	// If the payment is not settled, the method does nothing and returns `false`.
	// If it is, the method calls the contract's `paymentComplete` method and returns `true` (i.e. stop checking the payment).
	ProcessPaymentStatusResponse(request *bank.PaymentStatusResponse) (bool, error)
}

type MintRequestPayload struct {
	InstitutionId string `json:"institutionId"`
	SortCode      string `json:"sortCode"`
	AccountNumber string `json:"accountNumber"`
	Name          string `json:"name"`
	PublicKey     string `json:"publicKey"` // in base64
}

type AuthRequestPayload struct {
	Url       string `json:"url"`
	ConsentId string `json:"consentId"`
}

type AuthGrantedPayload struct {
	ConsentCode string `json:"consentCode"`
	PublicKey   string `json:"publicKey"` // in base64
}

type PendingPayment struct {
	RequestId string
	PaymentId string
}

func NewPaymentAuthRequest(envelope *contract.ProvableGBPMintRequest, payload *MintRequestPayload) bank.PaymentAuthRequest {

	tmpReq := hex.EncodeToString(envelope.RequestId[:])
	return bank.PaymentAuthRequest{
		RequestId:     tmpReq,
		InstitutionId: payload.InstitutionId,
		Amount:        ToDecimal(envelope.Amount, DECIMAL_DIGITS),
		Payer: bank.AccountDetails{
			Name:          payload.Name,
			SortCode:      payload.SortCode,
			AccountNumber: payload.AccountNumber,
		},
	}
}

// ToDecimal converts wei to decimal string
func ToDecimal(value *big.Int, decimalDigits int) string {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimalDigits)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result.String()
}

// ToWei converts decimal string to wei
func ToWei(am string, decimalDigits int) *big.Int {
	amount, _ := decimal.NewFromString(am)

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimalDigits)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}
