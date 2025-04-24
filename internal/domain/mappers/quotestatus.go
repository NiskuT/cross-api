package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
)

var quoteStatusToProto = map[valueobject.QuoteStatus]billingservice.QuoteStatus{
	valueobject.QuoteStatusUnspecified: billingservice.QuoteStatus_QUOTE_STATUS_UNSPECIFIED,
	valueobject.QuoteStatusPending:     billingservice.QuoteStatus_PENDING,
	valueobject.QuoteStatusApproved:    billingservice.QuoteStatus_APPROVED,
	valueobject.QuoteStatusRejected:    billingservice.QuoteStatus_REJECTED,
}

var protoToQuoteStatus = map[billingservice.QuoteStatus]valueobject.QuoteStatus{
	billingservice.QuoteStatus_QUOTE_STATUS_UNSPECIFIED: valueobject.QuoteStatusUnspecified,
	billingservice.QuoteStatus_PENDING:                  valueobject.QuoteStatusPending,
	billingservice.QuoteStatus_APPROVED:                 valueobject.QuoteStatusApproved,
	billingservice.QuoteStatus_REJECTED:                 valueobject.QuoteStatusRejected,
}

func QuoteStatusToProto(status valueobject.QuoteStatus) billingservice.QuoteStatus {
	if protoStatus, ok := quoteStatusToProto[status]; ok {
		return protoStatus
	}

	return billingservice.QuoteStatus_QUOTE_STATUS_UNSPECIFIED
}

func QuoteStatusFromProto(status billingservice.QuoteStatus) valueobject.QuoteStatus {
	if protoStatus, ok := protoToQuoteStatus[status]; ok {
		return protoStatus
	}

	return valueobject.QuoteStatusUnspecified
}
