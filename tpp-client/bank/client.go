package bank

import "crypto/tls"

// OpenBankingClient is an interface to be implemented by all OB client implementations.
type OpenBankingClient interface {

	// GetPaymentAuthAccessToken returns an access token for the given request ID.
	// The implementation may use any TPP identification method (MA-TLS, username/password,...).
	// It may also choose to cache the access token for a given request ID.
	GetPaymentAuthAccessToken(requestId string) (*AccessToken, error)

	// CreatePaymentAuthRequest returns a payment authorisation request, i.e. a pending consent.
	// The response contains the URL to be used by the payer to authorise the payment.
	CreatePaymentAuthRequest(payer *PaymentAuthRequest, access *AccessToken, beneficiary *AccountDetails) (*PaymentAuthResponse, error)

	// SubmitPayment submits a payment to the bank.
	// The supplied arguments must be identical to those provided in the `CreatePaymentAuthRequest` call.
	// If successful, the response contains the payment ID.
	// The payment ID is used to check the payment status, as payments may take some time to settle.
	SubmitPayment(data *PaymentAuthGranted, paymentAuthRequest *PaymentAuthRequest, beneficiary *AccountDetails) (*SubmitPaymentResponse, error)

	// GetPaymentStatus returns the status of a payment (settled, pending, failed).
	GetPaymentStatus(data *SubmitPaymentResponse) (*PaymentStatusResponse, error)
}

// PaymentAuthRequest contains the details for the payer
type PaymentAuthRequest struct {
	RequestId     string
	InstitutionId string
	Amount        string
	Payer         AccountDetails
}

// AccountDetails are the account details of a Payer or Beneficiary
type AccountDetails struct {
	SortCode      string
	AccountNumber string
	Name          string
}

// AccessToken wraps on Oauth2 token
type AccessToken struct {
	Token     string
	ExpiresIn int
}

// OauthClientCreds represents the various data items on Oauth client might use to identify itself
type OauthClientCreds struct {
	ClientId       string
	ClientSecret   string
	TransportCert  tls.Certificate
	SigningCert    tls.Certificate
	RedirectionUrl string
}

type PaymentAuthResponse struct {
	RequestId string
	Url       string
	ConsentId string
}

type PaymentAuthGranted struct {
	ConsentId   string
	RequestId   string
	ConsentCode string
}

type SubmitPaymentResponse struct {
	RequestId    string
	ConsentCode  string
	ConsentToken string
	PaymentId    string
}

type PaymentStatusResponse struct {
	RequestId string
	PaymentId string
	Status    string
	Settled   bool
}
