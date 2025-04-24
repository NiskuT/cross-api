package valueobject

type InvoiceItem struct {
	Description string  `json:"description,omitempty"`
	Quantity    int32   `json:"quantity,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	TaxRate     float64 `json:"tax_rate,omitempty"`
}

type QuoteItem struct {
	Description string  `json:"description,omitempty"`
	Quantity    int32   `json:"quantity,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	TaxRate     float64 `json:"tax_rate,omitempty"`
}
