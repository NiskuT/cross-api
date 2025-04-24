package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/repository"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidToken is returned when the token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")
)

type UserService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

type UserServiceConfiguration func(u *UserService) error

func NewUserService(cfgs ...UserServiceConfiguration) *UserService {
	impl := new(UserService)

	for _, cfg := range cfgs {
		if err := cfg(impl); err != nil {
			panic(err)
		}
	}

	return impl
}

func UserConfWithUserRepo(repo repository.UserRepository) UserServiceConfiguration {
	return func(u *UserService) error {
		u.userRepo = repo
		return nil
	}
}

func UserConfWithConfig(cfg *config.Config) UserServiceConfiguration {
	return func(u *UserService) error {
		u.cfg = cfg
		return nil
	}
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, email, password string) (*aggregate.JwtToken, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	return s.generateTokens(user)
}

// RefreshToken validates a refresh token and returns a new JWT token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*aggregate.JwtToken, error) {
	// Parse the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.Jwt.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Verify token type
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// Extract user ID
	var userID int32
	if id, ok := claims["sub"].(float64); ok {
		userID = int32(id)
	} else {
		return nil, ErrInvalidToken
	}

	// Get user from repository
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Generate new tokens
	return s.generateTokens(user)
}

// Helper function to generate JWT tokens
func (s *UserService) generateTokens(user *aggregate.User) (*aggregate.JwtToken, error) {
	roles := strings.Split(user.GetRole(), ",")

	// Create access token
	accessTokenClaims := jwt.MapClaims{
		"sub":   user.GetID(),
		"email": user.GetEmail(),
		"roles": roles,
		"iss":   "golene-evasion.com",
		"type":  "access",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.Jwt.SecretKey))
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshTokenClaims := jwt.MapClaims{
		"sub":  user.GetID(),
		"iss":  "golene-evasion.com",
		"type": "refresh",
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.Jwt.SecretKey))
	if err != nil {
		return nil, err
	}

	// Create token aggregate
	jwtToken := aggregate.NewJwtToken()
	jwtToken.SetAccessToken(accessTokenString)
	jwtToken.SetRefreshToken(refreshTokenString)

	return jwtToken, nil
}
