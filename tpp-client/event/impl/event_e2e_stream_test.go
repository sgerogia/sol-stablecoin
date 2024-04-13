package event_impl_test

// import (
// 	"encoding/base64"
// 	"encoding/json"
// 	"github.com/go-co-op/gocron"
// 	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
// 	bank_impl "github.com/sgerogia/sol-stablecoin/tpp-client/bank/impl"
// 	"github.com/sgerogia/sol-stablecoin/tpp-client/encrypt"
// 	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
// 	event_impl "github.com/sgerogia/sol-stablecoin/tpp-client/event/impl"
// 	"github.com/sgerogia/sol-stablecoin/tpp-client/schedule"
// 	test_util "github.com/sgerogia/sol-stablecoin/tpp-client/util"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/zap"
// 	"math/big"
// 	"testing"
// 	"time"
// )

// func TestEventHandlerLifecycleStreamEvents(t *testing.T) {

// 	// --- arrange ---

// 	// 1. Create handlers & clients
// 	sch := schedule.NewPaymentScheduler(testingCtx.l)

// 	bankClient := bank_impl.NewNatwestSandboxClient(
// 		30,
// 		&bank.OauthClientCreds{
// 			ClientId:       testingCtx.bankInfo.ClientId,
// 			ClientSecret:   testingCtx.bankInfo.ClientSecret,
// 			RedirectionUrl: testingCtx.bankInfo.RedirectUrl,
// 		},
// 		testingCtx.l,
// 	)

// 	handler := event_impl.NewEventHandler(
// 		testingCtx.chainInfo.TppContractClient,
// 		testingCtx.chainInfo.TppKeyPair,
// 		bankClient,
// 		test_util.Receiver(),
// 		sch,
// 		testingCtx.l)

// 	subscriber := event_impl.NewEventSubscriber(
// 		handler,
// 		testingCtx.chainInfo.TppContractClient,
// 		testingCtx.l)

// 	// task and schedule
// 	task := schedule.NewPaymentStatusTask(sch, bankClient, handler, testingCtx.l)
// 	s := gocron.NewScheduler(time.UTC)
// 	s.Every(5).Seconds().Do(task.CheckPaymentStatuses)
// 	s.StartAsync()
// 	defer s.Stop()

// 	payerClient := testingCtx.chainInfo.PayerContractClient
// 	payerKeyPair := testingCtx.chainInfo.PayerKeyPair

// 	payerEncKey := testingCtx.chainInfo.PayerKeyPair.PublicEncrKeyBytes()[:]

// 	// 2. Listen for events
// 	ok, err := subscriber.SubscribeToMintRequestEvent()
// 	assert.True(t, ok)
// 	require.NoError(t, err)

// 	ok, err = subscriber.SubscribeToAuthGrantedEvent()
// 	assert.True(t, ok)
// 	require.NoError(t, err)

// 	// --- act & assert ---

// 	// 0. Ensure payer has no GBP balance to begin with
// 	sess, err := payerClient.GetSingleUseSession()
// 	require.NoError(t, err)
// 	balance, err := sess.BalanceOf(sess.CallOpts.From)
// 	require.NoError(t, err)
// 	assert.Equal(t, int64(0), balance.Int64())

// 	// 1. Payer sends a mint request
// 	sess, err = payerClient.GetSingleUseSession()
// 	require.NoError(t, err)
// 	// ...get the TPP's public encryption key
// 	tppEncKey, err := sess.PublicKey()
// 	require.NoError(t, err)
// 	// ...create a mint request, make the call & mint a block
// 	data, err := json.Marshal(test_util.MintRequestPayload(payerEncKey))
// 	require.NoError(t, err)
// 	ethBox, err := payerKeyPair.Encrypt(data, (*[32]byte)(tppEncKey))
// 	require.NoError(t, err)
// 	encrData, err := json.Marshal(ethBox)
// 	require.NoError(t, err)
// 	amount := event.ToWei(test_util.AMOUNT, event.DECIMAL_DIGITS)
// 	_, err = sess.MintRequest(amount, encrData)
// 	require.NoError(t, err)
// 	testingCtx.chainInfo.Backend.Commit()

// 	// 2. TPP auto-handles the MintRequest event. Sleep while we are processing...
// 	time.Sleep(5 * time.Second)
// 	// ...Have we received the MintRequest event?
// 	logs := testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("MintRequest event").All()
// 	assert.Len(t, logs, 1)
// 	// ...Have we made the 2 OB calls?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("PaymentAuthAccess token request").All()
// 	assert.Len(t, logs, 1)
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("Payment auth request").All()
// 	assert.Len(t, logs, 1)
// 	// ...Have we called the contract with the URL?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessageSnippet("MintRequest processed. AuthRequest call").All()
// 	assert.Len(t, logs, 1)
// 	// ...If yes, mint a block to proceed
// 	testingCtx.chainInfo.Backend.Commit()

// 	// 3. Payer receives AuthRequest event
// 	// ...read the event as Payer & decrypt the payload
// 	msgIter, err := testingCtx.chainInfo.PayerContractClient.GetEventFilterer().FilterAuthRequest(nil, nil, nil)
// 	require.NoError(t, err)
// 	assert.True(t, msgIter.Next(), "No AuthRequest event received")
// 	var authReqData encrypt.EthSigUtilBox
// 	require.NoError(t, json.Unmarshal(msgIter.Event.AuthEncryptedData, &authReqData))
// 	decryptedData, err := payerKeyPair.Decrypt(&authReqData)
// 	require.NoError(t, err)
// 	var authReqPayload event.AuthRequestPayload
// 	require.NoError(t, json.Unmarshal(decryptedData, &authReqPayload))

// 	// 4. Programmatic approval of consent
// 	authResp := bank.PaymentAuthResponse{
// 		RequestId: string(msgIter.Event.RequestId[:]),
// 		Url:       authReqPayload.Url,
// 		ConsentId: authReqPayload.ConsentId,
// 	}
// 	authGranted, err := bankClient.(*bank_impl.NatwestSandboxClient).ApproveConsent(
// 		&authResp,
// 		test_util.GetNatwestSandboxInfo().CustomerUsername)
// 	require.NoError(t, err)

// 	// 5. Send an AuthGranted request
// 	// ...encrypt the payload
// 	authGrantedPayload := event.AuthGrantedPayload{
// 		ConsentCode: authGranted.ConsentCode,
// 		PublicKey:   base64.StdEncoding.EncodeToString([]byte(payerEncKey)),
// 	}
// 	data, err = json.Marshal(authGrantedPayload)
// 	require.NoError(t, err)
// 	ethBox, err = payerKeyPair.Encrypt(data, (*[32]byte)(tppEncKey))
// 	// ...make the contract call & mint a block
// 	encrData, err = json.Marshal(ethBox)
// 	require.NoError(t, err)
// 	sess, err = payerClient.GetSingleUseSession()
// 	require.NoError(t, err)
// 	_, err = sess.AuthGranted(msgIter.Event.RequestId, encrData)
// 	require.NoError(t, err)
// 	testingCtx.chainInfo.Backend.Commit()

// 	// 6. Auto-handle the AuthGranted event. Sleep while we are processing...
// 	time.Sleep(5 * time.Second)
// 	// ...Have we received the AuthGranted event?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("AuthGranted event").All()
// 	assert.Len(t, logs, 1)
// 	// ...Have we submitted the payment?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("Payment submitted").All()
// 	assert.Len(t, logs, 1)
// 	// ...and payment settlement scheduled?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("payment").FilterMessage("Scheduling payment").All()
// 	assert.Len(t, logs, 1)
// 	// ...making sure there was no duplicate scheduled event?
// 	logs = testingCtx.o.FilterLevelExact(zap.ErrorLevel).
// 		FilterFieldKey("reqId").FilterMessageSnippet("Duplicate payment request").All()
// 	assert.Len(t, logs, 0)
// 	// ...and that everything completed successfully?
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("AuthGranted processed").All()
// 	assert.Len(t, logs, 1)

// 	// 7. Auto-handle the payment settlement. Sleep while we are processing...
// 	time.Sleep(10 * time.Second)
// 	// ...Have we called contract paymentComplete? If yes, mint a block
// 	logs = testingCtx.o.FilterLevelExact(zap.InfoLevel).
// 		FilterFieldKey("reqId").FilterMessage("PaymentComplete call").All()
// 	assert.Len(t, logs, 1)
// 	testingCtx.chainInfo.Backend.Commit()
// 	// ...Have we minted the new tokens and updated the payer's balance minus seignorage?
// 	sess, err = payerClient.GetSingleUseSession()
// 	require.NoError(t, err)
// 	balance, err = sess.BalanceOf(sess.CallOpts.From)
// 	require.NoError(t, err)
// 	am := event.ToWei(test_util.AMOUNT, event.DECIMAL_DIGITS)
// 	netAm := big.NewInt(test_util.NET_OF_SEIGNORAGE)
// 	hundPerc := big.NewInt(test_util.ONE_HUND_PERC)
// 	assert.Equal(t, am.Mul(am, netAm).Div(am, hundPerc), balance)

// 	// 🍾🎉🎇🥂
// }
