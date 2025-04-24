package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
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
	// ErrEmailSendingFailed is returned when the system fails to send an email
	ErrEmailSendingFailed = errors.New("failed to send email")
	// ErrMissingEmailConfig is returned when email configuration is missing
	ErrMissingEmailConfig = errors.New("email configuration is missing")
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

// Helper function to generate a random password
func generateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Helper function to send an email
func (s *UserService) sendEmail(to, subject, body string) error {
	if s.cfg.Email.Host == "" {
		return ErrMissingEmailConfig
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", s.cfg.Email.Username, s.cfg.Email.Password, s.cfg.Email.Host)

	// Compose the email
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.cfg.Email.From, to, subject, body))

	// Connect to the server, authenticate, set the sender and recipient, and send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.cfg.Email.Host, s.cfg.Email.Port),
		auth,
		s.cfg.Email.From,
		[]string{to},
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// AddUserToCompetition adds the referee role for a specific competition to a user
func (s *UserService) AddUserToCompetition(ctx context.Context, email string, competitionID int32) error {
	// Get the user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Create the new role
	newRole := fmt.Sprintf("referee:%d", competitionID)

	user.AddRole(newRole)
	// Save the changes
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) SetUserAsAdmin(ctx context.Context, email string, competitionID int32) (*aggregate.JwtToken, error) {
	// Get the user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Set the user as admin
	newRole := fmt.Sprintf("admin:%d", competitionID)
	user.AddRole(newRole)

	// Save the changes
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	return s.generateTokens(user)
}

// InviteUser creates a new user with a referee role for a specific competition and sends an invitation email
func (s *UserService) InviteUser(ctx context.Context, firstName, lastName, email string, competitionID int32) error {
	// Check if the user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// User already exists, just add the competition role
		return s.AddUserToCompetition(ctx, email, competitionID)
	}

	// Generate a random password
	password := generateRandomPassword(12)

	// Create a new user
	user := aggregate.NewUser()
	user.SetEmail(email)
	user.SetFirstName(firstName)
	user.SetLastName(lastName)

	// Set referee role for the specified competition
	role := fmt.Sprintf("referee:%d", competitionID)
	user.SetRole(role)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.SetPasswordHash(string(hashedPassword))

	// Save the user to the database
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Prepare and send the invitation email
	subject := "Bienvenue à Cross Competition - Invitation d'Arbitre"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Bienvenue à Cross Competition !</h2>
			<p>Cher/Chère %s %s,</p>
			<p>Vous avez été invité(e) en tant qu'arbitre pour la compétition ID : %d.</p>
			<p>Voici vos identifiants de connexion :</p>
			<ul>
				<li><strong>Email :</strong> %s</li>
				<li><strong>Mot de passe :</strong> %s</li>
			</ul>
			<p>Veuillez vous connecter à notre plateforme et changer votre mot de passe après votre première connexion.</p>
			<p>Cordialement,<br>L'équipe Cross Competition</p>
		</body>
		</html>
	`, firstName, lastName, competitionID, email, password)

	err = s.sendEmail(email, subject, body)
	if err != nil {
		// Note: Even if email sending fails, the user has been created
		return fmt.Errorf("user created but email sending failed: %w", err)
	}

	return nil
}
