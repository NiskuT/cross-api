package models

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
)

type CreateInvoiceRequest struct {
	CustomerId    string                     `json:"customer_id,omitempty"`
	InvoiceNumber string                     `json:"invoice_number,omitempty"`
	InvoiceDate   int64                      `json:"invoice_date,omitempty"`
	DueDate       int64                      `json:"due_date,omitempty"`
	Items         []*valueobject.InvoiceItem `json:"items,omitempty"`
	QuoteId       string                     `json:"quote_id,omitempty"`
	Notes         string                     `json:"notes,omitempty"`
}

type UpdateInvoiceRequest struct {
	CustomerId    string                     `json:"customer_id,omitempty"`
	InvoiceNumber string                     `json:"invoice_number,omitempty"`
	InvoiceDate   int64                      `json:"invoice_date,omitempty"`
	DueDate       int64                      `json:"due_date,omitempty"`
	Items         []*valueobject.InvoiceItem `json:"items,omitempty"`
	QuoteId       string                     `json:"quote_id,omitempty"`
	Notes         string                     `json:"notes,omitempty"`
	PaymentStatus valueobject.PaymentStatus  `json:"payment_status,omitempty"`
}

type Invoice struct {
	InvoiceId     string                     `json:"invoice_id,omitempty"`
	InvoiceNumber string                     `json:"invoice_number,omitempty"`
	Customer      *Customer                  `json:"customer,omitempty"`
	InvoiceDate   int64                      `json:"invoice_date,omitempty"`
	DueDate       int64                      `json:"due_date,omitempty"`
	Items         []*valueobject.InvoiceItem `json:"items,omitempty"`
	QuoteId       string                     `json:"quote_id,omitempty"`
	Notes         string                     `json:"notes,omitempty"`
	PaymentStatus valueobject.PaymentStatus  `json:"payment_status,omitempty"`
	PaymentDate   int64                      `json:"payment_date,omitempty"`
	PaymentMethod string                     `json:"payment_method,omitempty"`
	Metadata      *valueobject.Metadata      `json:"metadata,omitempty"`
}

type Invoices struct {
	Invoices   []*Invoice          `json:"invoices,omitempty"`
	Pagination *commons.Pagination `json:"pagination,omitempty"`
}
