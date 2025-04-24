package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
)

func ToProtoCreateInvoiceRequest(req *models.CreateInvoiceRequest, userId string) billingservice.CreateInvoiceRequest {
	return billingservice.CreateInvoiceRequest{
		CustomerId:    req.CustomerId,
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   req.InvoiceDate,
		DueDate:       req.DueDate,
		Items:         InvoiceItemsToProto(req.Items),
		QuoteId:       req.QuoteId,
		Notes:         req.Notes,
		UserId:        userId,
	}
}

func ToProtoUpdateInvoiceRequest(req *models.UpdateInvoiceRequest, userId, invoiceId string) billingservice.UpdateInvoiceRequest {
	return billingservice.UpdateInvoiceRequest{
		InvoiceId:     invoiceId,
		UserId:        userId,
		CustomerId:    req.CustomerId,
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   req.InvoiceDate,
		DueDate:       req.DueDate,
		Items:         InvoiceItemsToProto(req.Items),
		QuoteId:       req.QuoteId,
		Notes:         req.Notes,
		PaymentStatus: PaymentStatusToProto(req.PaymentStatus),
	}
}

func InvoiceItemToProto(item *valueobject.InvoiceItem) *billingservice.InvoiceItem {
	return &billingservice.InvoiceItem{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TaxRate:     item.TaxRate,
	}
}

func InvoiceItemFromProto(item *billingservice.InvoiceItem) *valueobject.InvoiceItem {
	return &valueobject.InvoiceItem{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TaxRate:     item.TaxRate,
	}
}

func InvoiceItemsFromProto(items []*billingservice.InvoiceItem) []*valueobject.InvoiceItem {
	invoiceItems := make([]*valueobject.InvoiceItem, len(items))
	for i, item := range items {
		invoiceItems[i] = InvoiceItemFromProto(item)
	}
	return invoiceItems
}

func InvoiceItemsToProto(items []*valueobject.InvoiceItem) []*billingservice.InvoiceItem {
	invoiceItems := make([]*billingservice.InvoiceItem, len(items))
	for i, item := range items {
		invoiceItems[i] = InvoiceItemToProto(item)
	}
	return invoiceItems
}

func InvoiceFromProto(invoice *billingservice.Invoice, customer *models.Customer) *models.Invoice {
	return &models.Invoice{
		InvoiceId:     invoice.InvoiceId,
		InvoiceNumber: invoice.InvoiceNumber,
		Customer:      customer,
		InvoiceDate:   invoice.InvoiceDate,
		DueDate:       invoice.DueDate,
		Items:         InvoiceItemsFromProto(invoice.Items),
		QuoteId:       invoice.QuoteId,
		Notes:         invoice.Notes,
		PaymentStatus: PaymentStatusFromProto(invoice.PaymentStatus),
		PaymentDate:   invoice.PaymentDate,
		PaymentMethod: invoice.PaymentMethod,
		Metadata:      MetadataFromProto(invoice.Metadata),
	}
}

func InvoicesFromProto(invoices []*billingservice.Invoice, customers map[string]*models.Customer, pagination *commons.Pagination) *models.Invoices {
	invoiceModels := make([]*models.Invoice, len(invoices))
	for i, invoice := range invoices {
		customer, ok := customers[invoice.CustomerId]
		if !ok {
			customer = &models.Customer{
				ID: invoice.CustomerId,
			}
		}
		invoiceModels[i] = InvoiceFromProto(invoice, customer)
	}
	return &models.Invoices{
		Invoices:   invoiceModels,
		Pagination: pagination,
	}
}
