package valueobject

type PaymentStatus string

const (
	PaymentStatusUnspecified PaymentStatus = "PAYMENT_STATUS_UNSPECIFIED"
	PaymentStatusUnpaid      PaymentStatus = "UNPAID"
	PaymentStatusPaid        PaymentStatus = "PAID"
	PaymentStatusOverdue     PaymentStatus = "OVERDUE"
)
