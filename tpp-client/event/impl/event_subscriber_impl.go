package event_impl

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	event2 "github.com/ethereum/go-ethereum/event"
	"github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"go.uber.org/zap"
)

const (
	MINT_REQUEST = "MintRequest"
	AUTH_GRANTED = "AuthGranted"
)

type EventSubscriberImpl struct {
	handler       event.EventHandler
	contract      *contract.ContractClient
	subscriptions map[string]*event2.Subscription
	l             *zap.SugaredLogger
}

// NewEventSubscriber returns a naive implementation of EventSubscriber.
// The instance does not persist messages, nor does it have a dead letter queue.
func NewEventSubscriber(
	_handler event.EventHandler,
	_contract *contract.ContractClient,
	_l *zap.SugaredLogger) event.EventSubscriber {
	return &EventSubscriberImpl{
		handler:       _handler,
		contract:      _contract,
		subscriptions: make(map[string]*event2.Subscription),
		l:             _l}
}

func (s *EventSubscriberImpl) SubscribeToMintRequestEvent() (bool, error) {

	if s.subscriptions[MINT_REQUEST] != nil {
		s.l.Debugw("Already subscribed to MintRequest event")
		return false, nil
	}

	opts := bind.WatchOpts{
		Start:   nil,
		Context: context.Background(),
	}

	mrChan := make(chan *contract.ProvableGBPMintRequest)
	sub, err := s.contract.GetEventFilterer().WatchMintRequest(&opts, mrChan, nil, nil)
	if err != nil {
		s.l.Errorw("Error subscribing to MintRequest event", "error", err)
		return false, err
	}

	go mintRequestWorker(mrChan, &s.handler, sub, s.l)

	s.subscriptions[MINT_REQUEST] = &sub

	return true, nil
}

func mintRequestWorker(
	_mrChan chan *contract.ProvableGBPMintRequest,
	_h *event.EventHandler,
	sub event2.Subscription,
	_l *zap.SugaredLogger) {

	for {
		select {
		case mr := <-_mrChan:
			_l.Infow("Received MintRequest event", "event", mr)
			err := (*_h).ProcessMintRequest(mr)
			if err != nil {
				_l.Errorw("Error while processing MintRequest event", "error", err)
			}
		case err := <-sub.Err():
			_l.Errorw("Error while listening to MintRequest event", "error", err)
		}
	}
}

func (s *EventSubscriberImpl) UnsubscribeFromMintRequestEvent() bool {

	if s.subscriptions[MINT_REQUEST] == nil {
		s.l.Debugw("Not subscribed to MintRequest event")
		return false
	}
	sub := s.subscriptions[MINT_REQUEST]
	(*sub).Unsubscribe()
	delete(s.subscriptions, MINT_REQUEST)
	return true
}

func (s *EventSubscriberImpl) SubscribeToAuthGrantedEvent() (bool, error) {
	if s.subscriptions[AUTH_GRANTED] != nil {
		s.l.Debugw("Already subscribed to AuthGranted event")
		return false, nil
	}

	opts := bind.WatchOpts{
		Start:   nil,
		Context: context.Background(),
	}

	mrChan := make(chan *contract.ProvableGBPAuthGranted)
	sub, err := s.contract.GetEventFilterer().WatchAuthGranted(&opts, mrChan, nil, nil)
	if err != nil {
		s.l.Errorw("Error subscribing to AuthGranted event", "error", err)
		return false, err
	}

	go authGrantedWorker(mrChan, &s.handler, sub, s.l)

	s.subscriptions[AUTH_GRANTED] = &sub

	return true, nil
}

func authGrantedWorker(
	_mrChan chan *contract.ProvableGBPAuthGranted,
	_h *event.EventHandler,
	sub event2.Subscription,
	_l *zap.SugaredLogger) {

	for {
		select {
		case ag := <-_mrChan:
			_l.Infow("Received AuthGranted event", "event", ag)
			err := (*_h).ProcessAuthGranted(ag)
			if err != nil {
				_l.Errorw("Error while processing AuthGranted event", "error", err)
			}
		case err := <-sub.Err():
			_l.Errorw("Error while listening to AuthGranted event", "error", err)
		}
	}
}

func (s *EventSubscriberImpl) UnsubscribeFromAuthGrantedEvent() bool {
	if s.subscriptions[AUTH_GRANTED] == nil {
		s.l.Debugw("Not subscribed to AuthGranted event")
		return false
	}
	sub := s.subscriptions[AUTH_GRANTED]
	(*sub).Unsubscribe()
	delete(s.subscriptions, AUTH_GRANTED)
	return true
}
