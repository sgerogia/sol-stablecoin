package schedule

import (
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sgerogia/sol-stablecoin/tpp-client/contract"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"go.uber.org/zap"
)

type ContractEventTask interface {
	FetchAndProcessEvents()
}

type ContractEventTaskImpl struct {
	startingBlock  uint64
	contractClient *contract.ContractClient
	handler        *event.EventHandler
	l              *zap.SugaredLogger
}

func NewContractEventTask(
	_startFromBlock uint64,
	_contractClient *contract.ContractClient,
	_handler *event.EventHandler,
	_l *zap.SugaredLogger) ContractEventTask {

	return &ContractEventTaskImpl{
		startingBlock:  _startFromBlock,
		contractClient: _contractClient,
		handler:        _handler,
		l:              _l,
	}
}

/**
 * Fetches and processes events (MintRequest, AuthGranted) from the contract.
 */
func (t *ContractEventTaskImpl) FetchAndProcessEvents() {

	mintLB := t.fetchAndProcessMintRequests()
	authLB := t.fetchAndProcessAuthGranted()

	maxLB := uint64(math.Max(float64(mintLB), float64(authLB)))

	if maxLB > t.startingBlock {
		// FIXME: This is a hack. We actually need an internal tracker for event reqIds processed
		// We may get an incomplete event list for a block for whatever reason
		// Or we restart the task and need to catch up
		t.startingBlock = maxLB + 1
	}
}

/**
 * Fetches and processes MintRequest events from the contract.
 * @return the latest block number processed
 */
func (t *ContractEventTaskImpl) fetchAndProcessMintRequests() uint64 {

	// Fetch events from the contract
	filterOpts := bind.FilterOpts{
		Start: t.startingBlock,
		End:   nil,
	}

	events, err := (*t.contractClient).GetEventFilterer().FilterMintRequest(&filterOpts, nil, nil)
	if err != nil {
		t.l.Errorw("Error fetching events: " + err.Error())
		return t.startingBlock
	}

	var latestBlock uint64 = 0
	// Process the events
	for events.Next() {
		event := events.Event
		t.l.Infow("Processing event", "event", event.RequestId)
		// the actual processing of the event
		(*t.handler).ProcessMintRequest(event)
		latestBlock = event.Raw.BlockNumber
	}
	return latestBlock
}

/**
 * Fetches and processes AuthGranted events from the contract.
 * @return the latest block number processed
 */
func (t *ContractEventTaskImpl) fetchAndProcessAuthGranted() uint64 {

	// Fetch events from the contract
	filterOpts := bind.FilterOpts{
		Start: t.startingBlock,
		End:   nil,
	}

	events, err := (*t.contractClient).GetEventFilterer().FilterAuthGranted(&filterOpts, nil, nil)
	if err != nil {
		t.l.Errorw("Error fetching events: " + err.Error())
		return t.startingBlock
	}

	var latestBlock uint64 = 0
	// Process the events
	for events.Next() {
		event := events.Event
		t.l.Infow("Processing event", "event", event.RequestId)
		// the actual processing of the event
		(*t.handler).ProcessAuthGranted(event)
		latestBlock = event.Raw.BlockNumber
	}
	return latestBlock
}
