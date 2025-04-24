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

func getCustomerIDsFromQuotes(quotes []*billingservice.Quote) []string {
	customerIDs := make(map[string]bool)
	for _, quote := range quotes {
		customerIDs[quote.CustomerId] = true
	}
	customerIDsList := make([]string, 0, len(customerIDs))
	for customerID := range customerIDs {
		customerIDsList = append(customerIDsList, customerID)
	}
	return customerIDsList
}

// getQuote godoc
// @Summary      Returns a quote by ID.
// @Description  Returns a quote by ID with all the details.
// @Tags         quote
// @Accept       json
// @Produce      json
// @Param        id             path      string  true               "Quote ID"
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Success      200            {object}  models.Quote  		 				 "The quote object"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /quote/{id} [get]
func (s *Server) getQuote(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing quote ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	quote, err := s.quoteService.GetQuote(c.Request.Context(), &billingservice.GetQuoteRequest{QuoteId: id, UserId: user.Id})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), quote.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	quoteModel := mappers.QuoteFromProto(quote, customer)

	c.JSON(http.StatusOK, quoteModel)
}

// listQuote godoc
// @Summary      Returns a list of quotes.
// @Description  Returns a list of quotes with pagination.
// @Tags         quote
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true               					"Bearer <token>"
// @Param        page           query     int     false 						 					"Page number for pagination"
// @Param        limit          query     int     false 						 					"Number of events per page"
// @Success      200            {object}  models.Quotes	 											"The quote object list, and pagination details"
// @Failure      400            {object}  models.ErrorResponse       					"Bad Request"
// @Failure      401            {object}  models.ErrorResponse       					"Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       					"Internal Server Error"
// @Router       /quote [get]
func (s *Server) listQuote(c *gin.Context) {
	page, limit := getPagination(c)

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	req := &billingservice.ListQuotesRequest{UserId: user.Id, PageNumber: page, PageSize: limit}

	quotes, err := s.quoteService.ListQuotes(c.Request.Context(), req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customers, err := s.listCustomersByID(c.Request.Context(), user.Id, getCustomerIDsFromQuotes(quotes.Quotes))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, mappers.QuotesFromProto(quotes.Quotes, customers, quotes.Pagination))
}

// createQuote godoc
// @Summary      Creates a new quote.
// @Description  Creates a new quote with the provided details.
// @Tags         quote
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  													true    "Bearer <token>"
// @Param        quote          body      models.CreateQuoteRequest  				true  	"The quote object"
// @Success      201            {object}  models.Quote  		 												"The created quote object"
// @Failure      400            {object}  models.ErrorResponse       								"Bad Request"
// @Failure      401            {object}  models.ErrorResponse       								"Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       								"Internal Server Error"
// @Router       /quote [post]
func (s *Server) createQuote(c *gin.Context) {
	var req models.CreateQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	grpcRequest := mappers.ToProtoCreateQuoteRequest(&req, user.Id)

	quote, err := s.quoteService.CreateQuote(c.Request.Context(), &grpcRequest)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), quote.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	quoteModel := mappers.QuoteFromProto(quote, customer)

	c.JSON(http.StatusCreated, quoteModel)
}

// updateQuote godoc
// @Summary      Updates a quote.
// @Description  Updates a quote with the provided details.
// @Tags         quote
// @Accept       json
// @Produce      json
// @Param        id             path      string  													true  	"Quote ID"
// @Param        Authorization  header    string  													true    "Bearer <token>"
// @Param        quote          body      models.UpdateQuoteRequest  				true  	"The quote object"
// @Success      200            {object}  models.Quote  		 												"The updated quote object"
// @Failure      400            {object}  models.ErrorResponse       								"Bad Request"
// @Failure      401            {object}  models.ErrorResponse       								"Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       								"Internal Server Error"
// @Router       /quote/{id} [patch]
func (s *Server) updateQuote(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing quote ID"))
		return
	}

	var req models.UpdateQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	grpcRequest := mappers.ToProtoUpdateQuoteRequest(&req, id, user.Id)

	quote, err := s.quoteService.UpdateQuote(c.Request.Context(), &grpcRequest)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	customer, err := s.getCustomerModel(c.Request.Context(), quote.CustomerId, user.Id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	quoteModel := mappers.QuoteFromProto(quote, customer)

	c.JSON(http.StatusOK, quoteModel)
}

// deleteQuote godoc
// @Summary      Deletes a quote.
// @Description  Deletes a quote by ID.
// @Tags         quote
// @Accept       json
// @Produce      json
// @Param        id             path      string  true  				 "Quote ID"
// @Param        Authorization  header    string  true           "Bearer <token>"
// @Success      200            {object}  string   							 "Success message"
// @Failure      400            {object}  models.ErrorResponse   "Bad Request"
// @Failure      401            {object}  models.ErrorResponse   "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse   "Internal Server Error"
// @Router       /quote/{id} [delete]
func (s *Server) deleteQuote(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing quote ID"))
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	res, err := s.quoteService.DeleteQuote(c.Request.Context(), &billingservice.DeleteQuoteRequest{QuoteId: id, UserId: user.Id})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, res.Message)
}
