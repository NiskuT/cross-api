// @title           Orkys API
// @version         1.0
// @description     This is the API documentation for the Orkys API Gateway
// @termsOfService  http://example.com/terms/
//
// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  orkys.com@gmail.com
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host      localhost:9000
// @BasePath  /
package main

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/orkys/backend/gateway/internal/config"
	"gitlab.com/orkys/backend/gateway/internal/server"
	"gitlab.com/orkys/backend/gateway/internal/utils"
	billingservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/billing-service"
	customerservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/customer-service"
	userservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/user-service"
	"golang.org/x/net/context"
)

func main() {
	log.Info().Msg("Server is starting ...")

	app := &cobra.Command{
		Use:   "Orkys",
		Short: "Start and manage the Orkys service",
		Long:  "This command initializes and starts the Orkys service",
	}

	restCmd := &cobra.Command{
		Use:     "rest",
		Aliases: []string{"r"},
		Short:   "Start the Rest API",
		Long:    `This command initializes and starts the Orkys Rest API`,
		Run:     runRestServer,
	}

	app.AddCommand(restCmd)

	if err := app.Execute(); err != nil {
		log.Fatal()
	}
}

func runRestServer(_ *cobra.Command, _ []string) {
	log.Info().Msg("Starting the REST API server ...")
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("Loading configuration ...")
	cfg := config.New()

	log.Info().Msg("Creating gRPC clients ...")
	userServiceClient := utils.CreateGRPCClient(cfg.ServiceUri.UserService)
	userService := userservice.NewUserServiceClient(userServiceClient)
	authService := userservice.NewAuthServiceClient(userServiceClient)

	billingServiceClient := utils.CreateGRPCClient(cfg.ServiceUri.BillingService)
	invoiceService := billingservice.NewInvoiceServiceClient(billingServiceClient)
	quoteService := billingservice.NewQuoteServiceClient(billingServiceClient)

	customerServiceClient := utils.CreateGRPCClient(cfg.ServiceUri.CustomerService)
	customerService := customerservice.NewCustomerServiceClient(customerServiceClient)

	log.Info().Msg("Creating server ...")
	server, err := server.NewServer(
		server.ServerConfWithConfig(cfg),
		server.ServerConfWithUserService(userService),
		server.ServerConfWithAuthService(authService),
		server.ServerConfWithQuoteService(quoteService),
		server.ServerConfWithInvoiceService(invoiceService),
		server.ServerConfWithCustomerService(customerService),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	log.Info().Msg("Starting server ...")
	server.Start(cfg)
}
