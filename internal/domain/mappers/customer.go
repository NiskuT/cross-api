package mappers

import (
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	customerservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/customer-service"
)

func CreateCustomerRequestToProto(req *models.CreateCustomerRequest, userId string) *customerservice.CreateCustomerRequest {
	return &customerservice.CreateCustomerRequest{
		UserId:      userId,
		Name:        req.Name,
		SocialName:  req.SocialName,
		Description: req.Description,
		PictureUrl:  req.PictureUrl,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
	}
}

func UpdateCustomerRequestToProto(req *models.UpdateCustomerRequest, userId, customerId string) *customerservice.UpdateCustomerRequest {
	return &customerservice.UpdateCustomerRequest{
		Id:          customerId,
		UserId:      userId,
		Name:        req.Name,
		SocialName:  req.SocialName,
		Description: req.Description,
		PictureUrl:  req.PictureUrl,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
	}
}

func CustomerToProto(customer *models.Customer) *customerservice.Customer {
	return &customerservice.Customer{
		Id:          customer.ID,
		Name:        customer.Name,
		SocialName:  customer.SocialName,
		Description: customer.Description,
		PictureUrl:  customer.PictureUrl,
		Email:       customer.Email,
		Phone:       customer.Phone,
		Address:     customer.Address,
		PostalCode:  customer.PostalCode,
		Country:     customer.Country,
		Metadata:    MetadataToProto(&customer.Metadata),
	}
}

func CustomerFromProto(proto *customerservice.Customer) *models.Customer {
	return &models.Customer{
		ID:          proto.Id,
		Name:        proto.Name,
		SocialName:  proto.SocialName,
		Description: proto.Description,
		PictureUrl:  proto.PictureUrl,
		Email:       proto.Email,
		Phone:       proto.Phone,
		Address:     proto.Address,
		PostalCode:  proto.PostalCode,
		Country:     proto.Country,
		Metadata:    *MetadataFromProto(proto.Metadata),
	}
}
