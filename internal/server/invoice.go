package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/orkys/backend/gateway/internal/domain/mappers"
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/server/middlewares"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
)

func getCustomerIDsFromInvoices(invoices []*billingservice.Invoice) []string {
	customerIDs := make(map[string]bool)
	for _, invoice := range invoices {
		customerIDs[invoice.CustomerId] = true
	}
	customerIDsList := make([]string, 0, len(customerIDs))
	for customerID := range customerIDs {
		customerIDsList = append(customerIDsList, customerID)
	}
	return customerIDsList
}

// getInvoice godoc
// @Summary      Returns an invoice by ID.
// @Description  Returns an invoice by ID with all the details.
// @Tags         invoice
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               "Invoice ID"
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Success      200            {object}  models.Invoice          	 "The invoice object"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /invoice/{id} [get]
func (s *Server) getInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing invoice ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	invoice, err := s.invoiceService.GetInvoice(c.Request.Context(), &billingservice.GetInvoiceRequest{
		InvoiceId: id,
		UserId:    user.Id,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), invoice.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	invoiceResp := mappers.InvoiceFromProto(invoice, customer)

	c.JSON(http.StatusOK, invoiceResp)
}

// listInvoice godoc
// @Summary      Returns a list of invoices.
// @Description  Returns a list of invoices with pagination.
// @Tags         invoice
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Param        page           query     int     false              "Page number for pagination"
// @Param        limit          query     int     false              "Number of invoices per page"
// @Success      200            {object}  models.Invoices          	 "The invoice list and pagination details"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /invoice [get]
func (s *Server) listInvoice(c *gin.Context) {
	page, limit := getPagination(c)

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	req := &billingservice.ListInvoicesRequest{
		UserId:     user.Id,
		PageNumber: page,
		PageSize:   limit,
	}

	invoices, err := s.invoiceService.ListInvoices(c.Request.Context(), req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customers, err := s.listCustomersByID(c.Request.Context(), user.Id, getCustomerIDsFromInvoices(invoices.Invoices))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, mappers.InvoicesFromProto(invoices.Invoices, customers, invoices.Pagination))
}

// createInvoice godoc
// @Summary      Creates a new invoice.
// @Description  Creates a new invoice with the provided details.
// @Tags         invoice
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string               					true    	"Bearer <token>"
// @Param        invoice        body      models.CreateInvoiceRequest   true    	"The invoice object"
// @Success      201            {object}  models.Invoice          								"The created invoice object"
// @Failure      400            {object}  models.ErrorResponse       							"Bad Request"
// @Failure      401            {object}  models.ErrorResponse       							"Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       							"Internal Server Error"
// @Router       /invoice [post]
func (s *Server) createInvoice(c *gin.Context) {
	var req models.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	grpcRequest := mappers.ToProtoCreateInvoiceRequest(&req, user.Id)

	invoice, err := s.invoiceService.CreateInvoice(c.Request.Context(), &grpcRequest)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), invoice.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	invoiceResp := mappers.InvoiceFromProto(invoice, customer)

	c.JSON(http.StatusCreated, invoiceResp)
}

// updateInvoice godoc
// @Summary      Updates an invoice.
// @Description  Updates an invoice with the provided details.
// @Tags         invoice
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               					"Invoice ID"
// @Param        Authorization  header    string  true               					"Bearer <token>"
// @Param        invoice        body      models.UpdateInvoiceRequest   true  "The invoice object"
// @Success      200            {object}  models.Invoice          						"The updated invoice object"
// @Failure      400            {object}  models.ErrorResponse       					"Bad Request"
// @Failure      401            {object}  models.ErrorResponse       					"Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       					"Internal Server Error"
// @Router       /invoice/{id} [patch]
func (s *Server) updateInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing invoice ID"))
		return
	}

	var req models.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	grpcRequest := mappers.ToProtoUpdateInvoiceRequest(&req, user.Id, id)

	invoice, err := s.invoiceService.UpdateInvoice(c.Request.Context(), &grpcRequest)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), invoice.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	invoiceResp := mappers.InvoiceFromProto(invoice, customer)

	c.JSON(http.StatusOK, invoiceResp)
}

// deleteInvoice godoc
// @Summary      Deletes an invoice.
// @Description  Deletes an invoice by ID.
// @Tags         invoice
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               "Invoice ID"
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Success      200            {object}  string                 "Success message"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /invoice/{id} [delete]
func (s *Server) deleteInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing invoice ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	res, err := s.invoiceService.DeleteInvoice(c.Request.Context(), &billingservice.DeleteInvoiceRequest{
		InvoiceId: id,
		UserId:    user.Id,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, res.Message)
}
