package models

import "gitlab.com/orkys/backend/gateway/internal/domain/valueobject"

type Customer struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	SocialName  string               `json:"social_name"`
	Description string               `json:"description"`
	PictureUrl  string               `json:"picture_url"`
	Email       string               `json:"email"`
	Phone       string               `json:"phone"`
	Address     string               `json:"address"`
	PostalCode  string               `json:"postal_code"`
	Country     string               `json:"country"`
	Metadata    valueobject.Metadata `json:"metadata"`
}

type CreateCustomerRequest struct {
	Name        string `json:"name"`
	SocialName  string `json:"social_name"`
	Description string `json:"description"`
	PictureUrl  string `json:"picture_url"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}

type UpdateCustomerRequest struct {
	Name        string `json:"name"`
	SocialName  string `json:"social_name"`
	Description string `json:"description"`
	PictureUrl  string `json:"picture_url"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}
