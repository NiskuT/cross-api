package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
)

var paymentStatusToProto = map[valueobject.PaymentStatus]billingservice.PaymentStatus{
	valueobject.PaymentStatusUnspecified: billingservice.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
	valueobject.PaymentStatusUnpaid:      billingservice.PaymentStatus_UNPAID,
	valueobject.PaymentStatusPaid:        billingservice.PaymentStatus_PAID,
	valueobject.PaymentStatusOverdue:     billingservice.PaymentStatus_OVERDUE,
}

var protoToPaymentStatus = map[billingservice.PaymentStatus]valueobject.PaymentStatus{
	billingservice.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED: valueobject.PaymentStatusUnspecified,
	billingservice.PaymentStatus_UNPAID:                     valueobject.PaymentStatusUnpaid,
	billingservice.PaymentStatus_PAID:                       valueobject.PaymentStatusPaid,
	billingservice.PaymentStatus_OVERDUE:                    valueobject.PaymentStatusOverdue,
}

func PaymentStatusToProto(paymentStatus valueobject.PaymentStatus) billingservice.PaymentStatus {
	if status, ok := paymentStatusToProto[paymentStatus]; ok {
		return status
	}

	return billingservice.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
}

func PaymentStatusFromProto(paymentStatus billingservice.PaymentStatus) valueobject.PaymentStatus {
	if status, ok := protoToPaymentStatus[paymentStatus]; ok {
		return status
	}

	return valueobject.PaymentStatusUnspecified
}
