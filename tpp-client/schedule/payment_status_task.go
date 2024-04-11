package schedule

import (
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"go.uber.org/zap"
)

type PaymentStatusTask interface {
	CheckPaymentStatuses()
}

type PaymentStatusTaskImpl struct {
	scheduler  *PaymentStatusScheduler
	bankClient *bank.OpenBankingClient
	handler    *event.EventHandler
	l          *zap.SugaredLogger
}

func NewPaymentStatusTask(
	_scheduler PaymentStatusScheduler,
	_bankClient bank.OpenBankingClient,
	_handler event.EventHandler,
	_l *zap.SugaredLogger) PaymentStatusTask {

	return &PaymentStatusTaskImpl{
		scheduler:  &_scheduler,
		bankClient: &_bankClient,
		handler:    &_handler,
		l:          _l,
	}
}

func (t *PaymentStatusTaskImpl) CheckPaymentStatuses() {

	// TODO: an obvious perf. improvement would be to launch a goroutine for each payment
	// and wait for all of them to complete. This would be a good exercise for the reader.
	// Real-world payments take some time to settle, so this would improve things.
	// OTOH a parallelisation here would potentially be at odds with the current design of the contract client, which
	// assumes serial execution.
	for _, payment := range (*t.scheduler).GetScheduledPayments() {

		t.l.Infow("Checking payment status",
			"requestId", payment.RequestId,
			"payment", payment.PaymentId)

		status, err := (*t.bankClient).GetPaymentStatus(payment)

		// TODO: errors below should be introspected to decide what to do
		// Now we just log and move on (i.e. check again in next cycle)
		if err != nil {
			t.l.Errorw("Error getting payment status: "+err.Error(),
				"requestId", payment.RequestId,
				"payment", payment.PaymentId)
		} else {
			ok, err := (*t.handler).ProcessPaymentStatusResponse(status)
			// TODO: error below should be introspected to decide what to do (reschedule or remove?)
			if err != nil {
				t.l.Errorw("Error processing payment status response: "+err.Error(),
					"requestId", payment.RequestId,
					"payment", payment.PaymentId)
			}
			if ok {
				(*t.scheduler).UnschedulePayment(payment)
			}
		}
	}
}
