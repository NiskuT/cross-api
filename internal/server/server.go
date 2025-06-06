package server

import (
	"fmt"
	"time"

	_ "github.com/NiskuT/cross-api/docs"
	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/domain/service"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerConfiguration func(s *Server) error

type Server struct {
	conf               *config.Config
	userService        service.UserService
	competitionService service.CompetitionService
	runService         service.RunService
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

func ServerConfWithUserService(userService service.UserService) ServerConfiguration {
	return func(s *Server) error {
		s.userService = userService
		return nil
	}
}

func ServerConfWithConfig(conf *config.Config) ServerConfiguration {
	return func(s *Server) error {
		s.conf = conf
		return nil
	}
}

func ServerConfWithCompetitionService(competitionService service.CompetitionService) ServerConfiguration {
	return func(s *Server) error {
		s.competitionService = competitionService
		return nil
	}
}

func ServerConfWithRunService(runService service.RunService) ServerConfiguration {
	return func(s *Server) error {
		s.runService = runService
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

	middlewares.SecureMode = cfg.SecureMode

	router.MaxMultipartMemory = 5 << 30

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.PUT("/login", s.login)
	router.POST("/logout", s.logout)

	router.Use(middlewares.Authentication(cfg.Jwt.SecretKey, s.userService))

	router.POST("/competition", s.createCompetition)
	router.GET("/competition", s.listCompetitions)
	router.POST("/competition/zone", s.addZoneToCompetition)
	router.PUT("/competition/zone", s.updateZoneInCompetition)
	router.DELETE("/competition/zone", s.deleteZoneFromCompetition)
	router.POST("/competition/participants", s.addParticipantsToCompetition)
	router.POST("/competition/referee", s.addRefereeToCompetition)
	router.GET("/competition/:competitionID/participant/:dossard", s.getParticipant)
	router.GET("/competition/:competitionID/participants", s.listParticipantsByCategory)
	router.GET("/competition/:competitionID/zones", s.listZones)
	router.GET("/competition/:competitionID/liveranking", s.getLiveranking)
	router.POST("/participant", s.createParticipant)
	router.POST("/run", s.createRun)
	return router
}

func RespondError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, models.ErrorResponse{
		Code:    statusCode,
		Message: err.Error(),
	})
}
