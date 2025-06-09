package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	Port int
}

type Database struct {
	Name string
	Uri  string
}
type Jwt struct {
	SecretKey string
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type RateLimitConfig struct {
	LoginAttempts          int
	LoginWindow            time.Duration
	ForgotPasswordAttempts int
	ForgotPasswordWindow   time.Duration
}

type Config struct {
	Service      Service
	Database     Database
	ClientURI    string
	Jwt          Jwt
	AllowOrigins []string
	Email        EmailConfig
	SecureMode   bool
	RateLimit    RateLimitConfig
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

	c.Database.Name = getStringFromEnv("DB_NAME")
	c.Database.Uri = getStringFromEnv("DB_URI")

	c.Jwt.SecretKey = getStringFromEnv("JWT_SECRET_KEY")

	c.Email.Host = getStringFromEnv("EMAIL_HOST")
	c.Email.Port = getIntFromEnv("EMAIL_PORT")
	c.Email.Username = getStringFromEnv("EMAIL_USERNAME")
	c.Email.Password = getStringFromEnv("EMAIL_PASSWORD")
	c.Email.From = getStringFromEnv("EMAIL_FROM")

	// Rate limiting configuration with defaults
	c.RateLimit.LoginAttempts = getIntFromEnvWithDefault("LOGIN_RATE_LIMIT_ATTEMPTS", 5)
	c.RateLimit.LoginWindow = getDurationFromEnvWithDefault("LOGIN_RATE_LIMIT_WINDOW", 5*time.Minute)
	c.RateLimit.ForgotPasswordAttempts = getIntFromEnvWithDefault("FORGOT_PASSWORD_RATE_LIMIT_ATTEMPTS", 3)
	c.RateLimit.ForgotPasswordWindow = getDurationFromEnvWithDefault("FORGOT_PASSWORD_RATE_LIMIT_WINDOW", 1*time.Hour)

	// Origins
	origins := getStringFromEnv("ALLOW_ORIGINS")
	allowOrigins := strings.Split(origins, ",")
	c.AllowOrigins = allowOrigins

	c.SecureMode = getBoolFromEnv("SECURE_MODE")

	log.Info().Msgf("%s environment loaded successfully !", appEnv)
}

func (c *Config) GetEnv() string {
	return getStringFromEnv(AppEnv)
}

func getIntFromEnv(key string) int {
	myInt, err := strconv.Atoi(getStringFromEnv(key))
	if err != nil {
		panic(err)
	}

	return myInt
}

func getIntFromEnvWithDefault(key string, defaultValue int) int {
	valueStr := viper.GetString(key)
	if valueStr == "" {
		log.Info().Msgf("Environment variable %s not set, using default: %d", key, defaultValue)
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Warn().Msgf("Invalid value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

func getDurationFromEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	valueStr := viper.GetString(key)
	if valueStr == "" {
		log.Info().Msgf("Environment variable %s not set, using default: %s", key, defaultValue)
		return defaultValue
	}

	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Warn().Msgf("Invalid duration for %s: %s, using default: %s", key, valueStr, defaultValue)
		return defaultValue
	}

	return duration
}

func getStringFromEnv(key string) string {
	myString := viper.GetString(key)

	if myString == "" {
		log.Warn().Msgf("Environment variable %s is not set", key)
	}

	return myString
}

func getBoolFromEnv(key string) bool {
	myBool := viper.GetBool(key)

	if !myBool {
		log.Warn().Msgf("Environment variable %s is not set", key)
	}

	return myBool
}
