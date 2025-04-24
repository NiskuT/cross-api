package models

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
)

type CreateQuoteRequest struct {
	CustomerId string                   `json:"customer_id,omitempty"`
	QuoteDate  int64                    `json:"quote_date,omitempty"`
	ValidUntil int64                    `json:"valid_until,omitempty"`
	Items      []*valueobject.QuoteItem `json:"items,omitempty"`
	Discount   float64                  `json:"discount,omitempty"`
	Notes      string                   `json:"notes,omitempty"`
}

type UpdateQuoteRequest struct {
	CustomerId string                   `json:"customer_id,omitempty"`
	QuoteDate  int64                    `json:"quote_date,omitempty"`
	ValidUntil int64                    `json:"valid_until,omitempty"`
	Items      []*valueobject.QuoteItem `json:"items,omitempty"`
	Discount   float64                  `json:"discount,omitempty"`
	Notes      string                   `json:"notes,omitempty"`
	Status     valueobject.QuoteStatus  `json:"status,omitempty"`
}

type Quote struct {
	Id         string                   `json:"id,omitempty"`
	Customer   *Customer                `json:"customer,omitempty"`
	QuoteDate  int64                    `json:"quote_date,omitempty"`
	ValidUntil int64                    `json:"valid_until,omitempty"`
	Items      []*valueobject.QuoteItem `json:"items,omitempty"`
	Discount   float64                  `json:"discount,omitempty"`
	Notes      string                   `json:"notes,omitempty"`
	Status     valueobject.QuoteStatus  `json:"status,omitempty"`
	Metadata   *valueobject.Metadata    `json:"metadata,omitempty"`
}

type Quotes struct {
	Quotes     []*Quote            `json:"quotes,omitempty"`
	Pagination *commons.Pagination `json:"pagination,omitempty"`
}
