package schedule

import (
	"github.com/sgerogia/sol-stablecoin/tpp-client/bank"
	"go.uber.org/zap"
)

type PaymentStatusScheduler interface {
	SchedulePayment(payment *bank.SubmitPaymentResponse) bool
	GetScheduledPayments() []*bank.SubmitPaymentResponse
	UnschedulePayment(payment *bank.SubmitPaymentResponse) bool
}

type PaymentSchedulerImpl struct {
	payments map[string]*bank.SubmitPaymentResponse
	l        *zap.SugaredLogger
}

func NewPaymentScheduler(_l *zap.SugaredLogger) PaymentStatusScheduler {
	return &PaymentSchedulerImpl{
		payments: make(map[string]*bank.SubmitPaymentResponse),
		l:        _l}
}

// SchedulePayment adds a payment to the scheduler
func (t *PaymentSchedulerImpl) SchedulePayment(payment *bank.SubmitPaymentResponse) bool {

	t.l.Infow("Scheduling payment",
		"payment", payment.PaymentId)

	if t.payments[payment.RequestId] != nil {
		t.l.Infow("Payment already scheduled",
			"payment", payment.PaymentId)
		return false
	} else {
		t.payments[payment.RequestId] = payment
		return true
	}
}

// GetScheduledPayments returns a list of all scheduled payments
func (t *PaymentSchedulerImpl) GetScheduledPayments() []*bank.SubmitPaymentResponse {
	var payments []*bank.SubmitPaymentResponse
	for _, payment := range t.payments {
		payments = append(payments, payment)
	}
	return payments
}

// UnschedulePayment removes a payment from the scheduler
// Returns true if the payment was found and removed, false otherwise
func (t *PaymentSchedulerImpl) UnschedulePayment(payment *bank.SubmitPaymentResponse) bool {

	t.l.Infow("Unscheduling payment",
		"payment", payment.PaymentId)

	if t.payments[payment.RequestId] != nil {
		delete(t.payments, payment.RequestId)
		return true
	} else {
		t.l.Infow("Payment not scheduled",
			"payment", payment.PaymentId)
		return false
	}
}
