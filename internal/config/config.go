package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	AppEnv     = "APP_ENV"
	ServerPort = "SERVER_PORT"
)

type Environment string

const (
	Development Environment = "develop"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Local       Environment = "local"
	Test        Environment = "test"
)

type Service struct {
	Name string
	Port int
}

type Database struct {
	Name string
	Uri  string
}
type Jwt struct {
	SecretKey string
}

type Config struct {
	Service      Service
	Database     Database
	ClientURI    string
	Jwt          Jwt
	AllowOrigins []string
}

func New() *Config {
	c := new(Config)
	c.Load()

	return c
}

var TestEnvironments = map[Environment]Environment{
	Local:       Local,
	Development: Development,
	Test:        Test,
}

func (c *Config) Load() {
	viper.AutomaticEnv()

	appEnv := getStringFromEnv(AppEnv)

	_, isTestEnv := TestEnvironments[Environment(appEnv)]
	if isTestEnv {
		root, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		viper.SetConfigName(appEnv)
		viper.SetConfigType("env")
		viper.AddConfigPath(fmt.Sprintf("%s/internal/config", root))
		viper.AddConfigPath(fmt.Sprintf("..%s/internal/config", root))

		if err = viper.ReadInConfig(); err != nil {
			log.Error().Msgf("configuration file not found: %s", err.Error())
			log.Info().Msg("attempt to load configuration from environments variables")
		}
	}

	srvPort, err := strconv.Atoi(getStringFromEnv(ServerPort))
	if err != nil {
		panic(err)
	}

	c.Service.Port = srvPort
	c.Service.Name = getStringFromEnv("SERVICE_NAME")

	c.Jwt.SecretKey = getStringFromEnv("JWT_SECRET_KEY")

	// Origins
	origins := getStringFromEnv("ALLOW_ORIGINS")
	allowOrigins := strings.Split(origins, ",")
	c.AllowOrigins = allowOrigins

	log.Info().Msgf("%s environment loaded successfully !", appEnv)
}

func (c *Config) GetEnv() string {
	return getStringFromEnv(AppEnv)
}

func getStringFromEnv(key string) string {
	myString := viper.GetString(key)

	if myString == "" {
		log.Warn().Msgf("Environment variable %s is not set", key)
	}

	return myString
}
