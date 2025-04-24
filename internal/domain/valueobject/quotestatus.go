package valueobject

type QuoteStatus string

const (
	QuoteStatusUnspecified QuoteStatus = "QUOTE_STATUS_UNSPECIFIED"
	QuoteStatusPending     QuoteStatus = "PENDING"
	QuoteStatusApproved    QuoteStatus = "APPROVED"
	QuoteStatusRejected    QuoteStatus = "REJECTED"
)
