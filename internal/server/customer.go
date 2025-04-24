package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/orkys/backend/gateway/internal/domain/mappers"
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/server/middlewares"
	"gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons"
	customerservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/customer-service"
)

func (s *Server) getCustomerModel(ctx context.Context, customerId, userId string) (*models.Customer, error) {
	customer, err := s.customerService.GetCustomer(ctx, &customerservice.GetCustomerRequest{
		Id:     customerId,
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return mappers.CustomerFromProto(customer), nil
}

func (s *Server) listCustomersByID(ctx context.Context, userId string, allIds []string) (map[string]*models.Customer, error) {

	filters := []*commons.Filter{
		{
			Field: "id",
			Value: strings.Join(allIds, ","),
		},
	}

	customers, err := s.customerService.ListCustomers(ctx, &customerservice.ListCustomersRequest{
		UserId:  userId,
		Page:    1,
		Limit:   int32(len(allIds)),
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	customersMap := make(map[string]*models.Customer)
	for _, customer := range customers.Customers {
		customersMap[customer.Id] = mappers.CustomerFromProto(customer)
	}

	return customersMap, nil
}

// getCustomer godoc
// @Summary      Returns a customer by ID.
// @Description  Returns a customer by ID with all the details.
// @Tags         customer
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               "Customer ID"
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Success      200            {object}  models.Customer            "The customer object"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /customers/{id} [get]
func (s *Server) getCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing customer ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), id, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, customer)
}

// listCustomers godoc
// @Summary      Returns a list of customers.
// @Description  Returns a list of customers with pagination.
// @Tags         customer
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Param        page           query     int     false              "Page number for pagination"
// @Param        limit          query     int     false              "Number of customers per page"
// @Success      200            {object}  customerservice.ListCustomersResponse "The customer list and pagination details"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /customers [get]
func (s *Server) listCustomers(c *gin.Context) {
	page, limit := getPagination(c)

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	response, err := s.customerService.ListCustomers(c.Request.Context(), &customerservice.ListCustomersRequest{
		UserId: user.Id,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// createCustomer godoc
// @Summary      Creates a new customer.
// @Description  Creates a new customer with the provided details.
// @Tags         customer
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                        true    "Bearer <token>"
// @Param        customer      body      models.CreateCustomerRequest   true    "The customer object"
// @Success      201           {object}  models.Customer    										"The created customer object"
// @Failure      400           {object}  models.ErrorResponse       						"Bad Request"
// @Failure      401           {object}  models.ErrorResponse       						"Unauthorized"
// @Failure      500           {object}  models.ErrorResponse       						"Internal Server Error"
// @Router       /customers [post]
func (s *Server) createCustomer(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	customer, err := s.customerService.CreateCustomer(c.Request.Context(), mappers.CreateCustomerRequestToProto(&req, user.Id))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, mappers.CustomerFromProto(customer))
}

// updateCustomer godoc
// @Summary      Updates a customer.
// @Description  Updates a customer with the provided details.
// @Tags         customer
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               					"Customer ID"
// @Param        Authorization  header    string  true               					"Bearer <token>"
// @Param        customer      body      models.UpdateCustomerRequest   true  "The customer object"
// @Success      200           {object}  models.Customer    									"The updated customer object"
// @Failure      400           {object}  models.ErrorResponse       					"Bad Request"
// @Failure      401           {object}  models.ErrorResponse       					"Unauthorized"
// @Failure      500           {object}  models.ErrorResponse       					"Internal Server Error"
// @Router       /customers/{id} [patch]
func (s *Server) updateCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing customer ID"))
		return
	}

	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	customer, err := s.customerService.UpdateCustomer(c.Request.Context(), mappers.UpdateCustomerRequestToProto(&req, user.Id, id))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, mappers.CustomerFromProto(customer))
}

// deleteCustomer godoc
// @Summary      Deletes a customer.
// @Description  Deletes a customer by ID.
// @Tags         customer
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               "Customer ID"
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Success      200            {object}  commons.Response          "Success message"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /customers/{id} [delete]
func (s *Server) deleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing customer ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	response, err := s.customerService.DeleteCustomer(c.Request.Context(), &customerservice.DeleteCustomerRequest{
		Id:     id,
		UserId: user.Id,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
