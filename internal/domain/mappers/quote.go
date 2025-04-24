package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/domain/valueobject"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
)

func ToProtoCreateQuoteRequest(req *models.CreateQuoteRequest, userId string) billingservice.CreateQuoteRequest {
	return billingservice.CreateQuoteRequest{
		CustomerId: req.CustomerId,
		QuoteDate:  req.QuoteDate,
		ValidUntil: req.ValidUntil,
		Items:      QuoteItemsToProto(req.Items),
		Discount:   req.Discount,
		Notes:      req.Notes,
		UserId:     userId,
	}
}

func ToProtoUpdateQuoteRequest(req *models.UpdateQuoteRequest, quoteId, userId string) billingservice.UpdateQuoteRequest {
	return billingservice.UpdateQuoteRequest{
		QuoteId:    quoteId,
		CustomerId: req.CustomerId,
		QuoteDate:  req.QuoteDate,
		ValidUntil: req.ValidUntil,
		Items:      QuoteItemsToProto(req.Items),
		Discount:   req.Discount,
		Notes:      req.Notes,
		Status:     QuoteStatusToProto(req.Status),
		UserId:     userId,
	}
}

func QuoteItemToProto(item *valueobject.QuoteItem) *billingservice.QuoteItem {
	return &billingservice.QuoteItem{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TaxRate:     item.TaxRate,
	}
}

func QuoteItemsToProto(items []*valueobject.QuoteItem) []*billingservice.QuoteItem {
	protoItems := make([]*billingservice.QuoteItem, len(items))
	for i, item := range items {
		protoItems[i] = QuoteItemToProto(item)
	}
	return protoItems
}

func QuoteItemFromProto(item *billingservice.QuoteItem) *valueobject.QuoteItem {
	return &valueobject.QuoteItem{
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TaxRate:     item.TaxRate,
	}
}

func QuoteItemsFromProto(items []*billingservice.QuoteItem) []*valueobject.QuoteItem {
	quoteItems := make([]*valueobject.QuoteItem, len(items))
	for i, item := range items {
		quoteItems[i] = QuoteItemFromProto(item)
	}
	return quoteItems
}

func QuoteFromProto(quote *billingservice.Quote, customer *models.Customer) *models.Quote {
	return &models.Quote{
		Id:         quote.QuoteId,
		Customer:   customer,
		QuoteDate:  quote.QuoteDate,
		ValidUntil: quote.ValidUntil,
		Items:      QuoteItemsFromProto(quote.Items),
		Discount:   quote.Discount,
		Notes:      quote.Notes,
		Status:     QuoteStatusFromProto(quote.Status),
		Metadata:   MetadataFromProto(quote.Metadata),
	}
}

func QuotesFromProto(quotes []*billingservice.Quote, customers map[string]*models.Customer, pagination *commons.Pagination) *models.Quotes {
	quoteModels := make([]*models.Quote, len(quotes))
	for i, quote := range quotes {
		customer, ok := customers[quote.CustomerId]
		if !ok {
			customer = &models.Customer{
				ID: quote.CustomerId,
			}
		}
		quoteModels[i] = QuoteFromProto(quote, customer)
	}
	return &models.Quotes{
		Quotes:     quoteModels,
		Pagination: pagination,
	}
}
