package server

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "gitlab.com/orkys/backend/gateway/docs"
	"gitlab.com/orkys/backend/gateway/internal/config"
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/server/middlewares"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
	customerservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/customer-service"
	userservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/user-service"
)

type ServerConfiguration func(s *Server) error

type Server struct {
	conf            *config.Config
	userService     userservice.UserServiceClient
	authService     userservice.AuthServiceClient
	quoteService    billingservice.QuoteServiceClient
	invoiceService  billingservice.InvoiceServiceClient
	customerService customerservice.CustomerServiceClient
}

func NewServer(configs ...ServerConfiguration) (*Server, error) {
	s := &Server{}
	for _, config := range configs {
		if err := config(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func ServerConfWithUserService(userService userservice.UserServiceClient) ServerConfiguration {
	return func(s *Server) error {
		s.userService = userService
		return nil
	}
}

func ServerConfWithAuthService(authService userservice.AuthServiceClient) ServerConfiguration {
	return func(s *Server) error {
		s.authService = authService
		return nil
	}
}

func ServerConfWithQuoteService(quoteService billingservice.QuoteServiceClient) ServerConfiguration {
	return func(s *Server) error {
		s.quoteService = quoteService
		return nil
	}
}

func ServerConfWithInvoiceService(invoiceService billingservice.InvoiceServiceClient) ServerConfiguration {
	return func(s *Server) error {
		s.invoiceService = invoiceService
		return nil
	}
}

func ServerConfWithCustomerService(customerService customerservice.CustomerServiceClient) ServerConfiguration {
	return func(s *Server) error {
		s.customerService = customerService
		return nil
	}
}

func ServerConfWithConfig(conf *config.Config) ServerConfiguration {
	return func(s *Server) error {
		s.conf = conf
		return nil
	}
}

func (s *Server) Start(cfg *config.Config) {
	router := s.getRouter(cfg)
	err := router.Run(fmt.Sprintf(":%d", cfg.Service.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func (s *Server) getRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	if cfg.GetEnv() == string(config.Production) {
		gin.SetMode(gin.ReleaseMode)

		router = gin.New()
		router.Use(gin.Recovery())
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.MaxMultipartMemory = 5 << 30

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/login", s.login)
	router.POST("/register", s.register)
	router.GET("/refresh", s.refresh)

	router.GET("/oauth/:provider/initiate", s.oauthInitiate)
	router.GET("/oauth/:provider/callback", s.oauthCallback)

	router.Use(middlewares.Authentication(cfg.Jwt.SecretKey))

	router.GET("/list", s.list)
	router.PATCH("/user", s.updateUser)

	// quote
	router.GET("/quote/:id", s.getQuote)
	router.GET("/quote", s.listQuote)
	router.POST("/quote", s.createQuote)
	router.PATCH("/quote/:id", s.updateQuote)
	router.DELETE("/quote/:id", s.deleteQuote)

	// invoice
	router.GET("/invoice/:id", s.getInvoice)
	router.GET("/invoice", s.listInvoice)
	router.POST("/invoice", s.createInvoice)
	router.PATCH("/invoice/:id", s.updateInvoice)
	router.DELETE("/invoice/:id", s.deleteInvoice)

	// customer
	router.POST("/customers", s.createCustomer)
	router.GET("/customers", s.listCustomers)
	router.GET("/customers/:id", s.getCustomer)
	router.PATCH("/customers/:id", s.updateCustomer)
	router.DELETE("/customers/:id", s.deleteCustomer)

	return router
}

func RespondError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, models.ErrorResponse{
		Code:    statusCode,
		Message: err.Error(),
	})
}
