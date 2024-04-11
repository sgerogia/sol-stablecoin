package event_impl

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	"github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/sgerogia/sol-stablecoin/tpp-client/encrypt"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"github.com/sgerogia/sol-stablecoin/tpp-client/schedule"
	"go.uber.org/zap"
)

// EventHandlerImpl non-persistent implementation the EventHandler interface.
// Uses an in-memory map to track ongoing payment requests.
type EventHandlerImpl struct {
	contract    *contract.ContractClient
	keyPair     *encrypt.KeyPair
	bankClient  *bank.OpenBankingClient
	beneficiary *bank.AccountDetails
	scheduler   *schedule.PaymentStatusScheduler
	l           *zap.SugaredLogger
	cache       map[string]*ongoingRequest
}

type ongoingRequest struct {
	ConsentId          string
	PaymentAuthRequest *bank.PaymentAuthRequest
}

func NewEventHandler(
	_contract *contract.ContractClient,
	_keyPair *encrypt.KeyPair,
	_bankClient bank.OpenBankingClient,
	_beneficiary *bank.AccountDetails,
	_scheduler schedule.PaymentStatusScheduler,
	_l *zap.SugaredLogger) event.EventHandler {

	return &EventHandlerImpl{
		contract:    _contract,
		keyPair:     _keyPair,
		bankClient:  &_bankClient,
		beneficiary: _beneficiary,
		scheduler:   &_scheduler,
		l:           _l,
		cache:       make(map[string]*ongoingRequest),
	}
}

func (h *EventHandlerImpl) ProcessMintRequest(request *contract.ProvableGBPMintRequest) error {

	// --- unpack ETH event ---

	reqIdStr := hex.EncodeToString(request.RequestId[:])

	h.l.Infow("MintRequest event",
		"reqId", reqIdStr)

	// encrypted data
	var box encrypt.EthSigUtilBox
	if err := json.Unmarshal(request.EncryptedData, &box); err != nil {
		return errors.New("Error unmarshalling encr. data: " + err.Error())
	}
	// decrypt
	decr, err := h.keyPair.Decrypt(&box)
	if err != nil {
		return errors.New("Error decrypting encr. data: " + err.Error())
	}
	// recreate mintRequestPayload
	var mintRequestPayload event.MintRequestPayload
	if err = json.Unmarshal(decr, &mintRequestPayload); err != nil {
		return errors.New("Error unmarshalling MintRequest mintRequestPayload: " + err.Error())
	}

	// This debugging line doxxes all the encrypted info, but hey!
	h.l.Debugw("MintRequest payload",
		"reqId", reqIdStr,
		"payload", mintRequestPayload)

	// --- OpenBanking call ---

	token, err := (*h.bankClient).GetPaymentAuthAccessToken(reqIdStr)
	if err != nil {
		return err
	}

	var pAuthReq = event.NewPaymentAuthRequest(request, &mintRequestPayload)
	resp, err := (*h.bankClient).CreatePaymentAuthRequest(&pAuthReq, token, h.beneficiary)
	if err != nil {
		return err
	}

	authRequestPayload := event.AuthRequestPayload{
		Url:       resp.Url,
		ConsentId: resp.ConsentId,
	}
	arJson, err := json.Marshal(authRequestPayload)
	if err != nil {
		return errors.New("Error marshalling AuthRequestPayload: " + err.Error())
	}

	// --- Contract callback ---

	// recover their base64 public key
	publicKey, err := base64.StdEncoding.DecodeString(mintRequestPayload.PublicKey)
	if err != nil {
		return errors.New("Error decoding their public key: " + err.Error())
	}

	// encrypt response with their key
	authReqEncr, err := h.keyPair.Encrypt([]byte(arJson), (*[32]byte)(publicKey))
	if err != nil {
		return err
	}

	authReqEncrJson, err := json.Marshal(authReqEncr)
	if err != nil {
		return errors.New("Error marshalling encryption structure for AuthRequest: " + err.Error())
	}

	sess, err := h.contract.GetSingleUseSession()
	if err != nil {
		return err
	}

	tx, err := sess.AuthRequest(request.RequestId, []byte(authReqEncrJson))
	if err != nil {
		return errors.New("Error calling AuthRequest: " + err.Error())
	}

	// add to cache
	h.cache[reqIdStr] = &ongoingRequest{
		ConsentId:          resp.ConsentId,
		PaymentAuthRequest: &pAuthReq,
	}

	h.l.Infow("MintRequest processed. AuthRequest call",
		"reqId", reqIdStr,
		"txHash", tx.Hash().Hex())

	return nil
}

func (h *EventHandlerImpl) ProcessAuthGranted(request *contract.ProvableGBPAuthGranted) error {

	// --- unpack ETH event ---

	reqIdStr := hex.EncodeToString(request.RequestId[:])

	h.l.Infow("AuthGranted event",
		"reqId", reqIdStr)

	// encrypted data
	var box encrypt.EthSigUtilBox
	if err := json.Unmarshal(request.GrantEncryptedData, &box); err != nil {
		return errors.New("Error unmarshalling encr. data: " + err.Error())
	}
	// decrypt
	decr, err := h.keyPair.Decrypt(&box)
	if err != nil {
		return errors.New("Error decrypting encr. data: " + err.Error())
	}
	// recreate authGrantedPayload
	var authGrantedPayload event.AuthGrantedPayload
	if err = json.Unmarshal(decr, &authGrantedPayload); err != nil {
		return errors.New("Error unmarshalling AuthGranted payload: " + err.Error())
	}

	// --- OpenBanking call ---

	ongoingReq := h.cache[reqIdStr]
	if ongoingReq == nil {
		return errors.New("No ongoing request found for requestId: " + reqIdStr)
	}

	pAuthGranted := bank.PaymentAuthGranted{
		RequestId:   reqIdStr,
		ConsentId:   ongoingReq.ConsentId,
		ConsentCode: authGrantedPayload.ConsentCode,
	}
	resp, err := (*h.bankClient).SubmitPayment(&pAuthGranted, ongoingReq.PaymentAuthRequest, h.beneficiary)
	if err != nil {
		return err
	}
	h.l.Infow("Payment submitted",
		"reqId", reqIdStr,
		"paymentId", resp.PaymentId)

	scheduled := (*h.scheduler).SchedulePayment(resp)
	if !scheduled {
		// TODO we should really be doing more than logging
		// A duplicate request means a fundamental error somewhere
		h.l.Errorw("Duplicate payment request. Need to investigate!",
			"reqId", reqIdStr,
			"paymentId", resp.PaymentId)
	}

	h.l.Infow("AuthGranted processed",
		"reqId", reqIdStr,
		"paymentId", resp.PaymentId)

	return nil
}

func (h *EventHandlerImpl) ProcessPaymentStatusResponse(request *bank.PaymentStatusResponse) (bool, error) {

	h.l.Infow("Payment Status event",
		"reqId", request.RequestId,
		"paymentId", request.PaymentId,
		"settled", request.Settled)

	if request.Settled {
		sess, err := h.contract.GetSingleUseSession()
		if err != nil {
			return false, err
		}

		// go from hex back to bytes
		var reqId [32]byte
		tmp, err := hex.DecodeString(request.RequestId)
		if err != nil {
			return false, errors.New("Error converting requestId to bytes: " + err.Error())
		}
		copy(reqId[:], tmp)
		tx, err := sess.PaymentComplete(reqId)
		if err != nil {
			return false, errors.New("Error calling PaymentComplete: " + err.Error())
		}

		h.l.Infow("PaymentComplete call",
			"reqId", request.RequestId,
			"paymentId", request.PaymentId,
			"txHash", tx.Hash().Hex())
	}

	return request.Settled, nil
}
