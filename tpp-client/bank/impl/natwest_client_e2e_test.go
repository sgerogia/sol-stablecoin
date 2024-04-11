package bank_impl_test

import (
	"github.com/google/uuid"
	"github.com/sgerogia/hello-stablecoin/tpp-client/bank"
	bank_impl "github.com/sgerogia/hello-stablecoin/tpp-client/bank/impl"
	test_util "github.com/sgerogia/hello-stablecoin/tpp-client/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

// TestNatwestClientE2E tests the end-to-end flow of the Natwest client.
// It requires the following env vars to be set, otherwise execution is skipped:
// - NATWEST_SANDBOX_CLIENT_ID: the client id of the Natwest sandbox app
// - NATWEST_SANDBOX_CLIENT_SECRET: the client secret of the Natwest sandbox app
// - NATWEST_SANDBOX_REDIRECT_URL: the redirect url of the Natwest sandbox app
// - NATWEST_SANDBOX_CUSTOMER_USERNAME: the username of the customer to use for the consent flow (format: CUSTOMER_ID_OR_CUSTOMER_NUMBER@domainof.your.app)
func TestNatwestSandboxClient_E2E(t *testing.T) {

	// skip if env vars are not there
	info := test_util.GetNatwestSandboxInfo()
	if info == nil {
		t.Skip("Skipping test, env vars not set")
	}

	// arrange
	creds := bank.OauthClientCreds{
		ClientId:       info.ClientId,
		ClientSecret:   info.ClientSecret,
		RedirectionUrl: info.RedirectUrl,
	}
	l := zap.NewExample().Sugar()
	client := bank_impl.NewNatwestSandboxClient(30, &creds, l)
	reqId := uuid.New().String()
	payer := test_util.Payer()
	receiver := test_util.Receiver()
	customerUsername := info.CustomerUsername

	// --- act & assert ---

	// 1) get access token
	accessToken, err := client.GetPaymentAuthAccessToken(reqId)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken.Token)
	assert.Positive(t, accessToken.ExpiresIn)

	// 2) create payment auth request
	paymentAuthRequest := bank.PaymentAuthRequest{
		RequestId:     reqId,
		InstitutionId: test_util.INSTITUTION,
		Amount:        test_util.AMOUNT,
		Payer:         *payer,
	}
	authResp, err := client.CreatePaymentAuthRequest(&paymentAuthRequest, accessToken, receiver)
	require.NoError(t, err)
	assert.NotEmpty(t, authResp.Url)
	assert.NotEmpty(t, authResp.RequestId)

	// 3) approve consent (with hard type conversion)
	authGranted, err := client.(*bank_impl.NatwestSandboxClient).ApproveConsent(authResp, customerUsername)
	require.NoError(t, err)
	assert.NotEmpty(t, authGranted.ConsentCode)
	assert.NotEmpty(t, authGranted.RequestId)

	// 4) submit payment
	paymResp, err := client.SubmitPayment(authGranted, &paymentAuthRequest, receiver)
	require.NoError(t, err)
	assert.NotEmpty(t, paymResp.RequestId)
	assert.NotEmpty(t, paymResp.PaymentId)
	assert.NotEmpty(t, paymResp.ConsentToken)
	assert.NotEmpty(t, paymResp.ConsentCode)

	// 5) check payment
	// repeat 10 times with 1 sec sleep until status is "AcceptedSettlementCompleted"
	n := 0
	for n < 10 {
		checkResp, err := client.GetPaymentStatus(paymResp)
		require.NoError(t, err)
		if checkResp.Settled {
			assert.Equal(t, bank_impl.NATWEST_SETTLED_STATUS, checkResp.Status)
			break
		}
		time.Sleep(1 * time.Second)
		n++
	}
	assert.NotEqual(t, 10, n, "Payment not settled after 10 seconds")
}
