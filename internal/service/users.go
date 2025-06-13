package service

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// ErrMaximumRolesReached is returned when the user has reached the maximum number of roles
	ErrMaximumRolesReached = errors.New("maximum number of roles reached, contact support")
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
		log.Println("Error getting user by email:", err)
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(password)); err != nil {
		log.Println("Error comparing password:", err)
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
	roles := strings.Split(user.GetRoles(), ",")

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
	jwtToken.SetRoles(roles)

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
func (s *UserService) AddUserToCompetition(ctx context.Context, email string, competition *aggregate.Competition) error {
	// Get the user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Create the new role
	newRole := fmt.Sprintf("referee:%d", competition.GetID())

	user.AddRole(newRole)
	if len(user.GetRoles()) >= 500 {
		return ErrMaximumRolesReached
	}

	// Save the changes
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	// Send notification email to existing user
	subject := "Golene Evasion - Nouvelle Invitation d'Arbitre"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Nouvelle Invitation - Golene Evasion</h2>
			<p>Cher/Chère %s %s,</p>
			<p>Vous avez été invité(e) en tant qu'arbitre pour une nouvelle compétition : %s.</p>
			<p>Vous pouvez vous connecter avec vos identifiants existants sur <a href="https://cross.golene-evasion.com">notre plateforme</a>.</p>
			<p>Si vous avez oublié votre mot de passe, utilisez la fonction "Mot de passe oublié" sur la page de connexion.</p>
			<p>Cordialement,<br>L'équipe Golene Evasion</p>
		</body>
		</html>
	`, user.GetFirstName(), user.GetLastName(), competition.GetName())

	err = s.sendEmail(email, subject, body)
	if err != nil {
		// Note: Even if email sending fails, the role has been added successfully
		return fmt.Errorf("referee role added but email notification failed: %w", err)
	}

	return nil
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
func (s *UserService) InviteUser(ctx context.Context, firstName, lastName, email string, competition *aggregate.Competition) error {
	// Check if the user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// User already exists, just add the competition role
		return s.AddUserToCompetition(ctx, email, competition)
	}

	// Generate a random password
	password := generateRandomPassword(12)

	// Create a new user
	user := aggregate.NewUser()
	user.SetEmail(email)
	user.SetFirstName(firstName)
	user.SetLastName(lastName)

	// Set referee role for the specified competition
	role := fmt.Sprintf("referee:%d", competition.GetID())
	user.SetRoles(role)

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
	subject := "Bienvenue à Golene Evasion - Invitation d'Arbitre"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Bienvenue à Golene Evasion !</h2>
			<p>Cher/Chère %s %s,</p>
			<p>Vous avez été invité(e) en tant qu'arbitre pour la compétition : %s.</p>
			<p>Voici vos identifiants de connexion :</p>
			<ul>
				<li><strong>Email :</strong> %s</li>
				<li><strong>Mot de passe :</strong> %s</li>
			</ul>
			<p>Veuillez vous connecter à notre plateforme <a href="https://cross.golene-evasion.com">ici</a> et changer votre mot de passe après votre première connexion.</p>
			<p>Cordialement,<br>L'équipe Golene Evasion</p>
		</body>
		</html>
	`, firstName, lastName, competition.GetName(), email, password)

	err = s.sendEmail(email, subject, body)
	if err != nil {
		// Note: Even if email sending fails, the user has been created
		return fmt.Errorf("user created but email sending failed: %w", err)
	}

	return nil
}

// ChangePassword allows a user to change their password by verifying their current password
func (s *UserService) ChangePassword(ctx context.Context, userID int32, currentPassword, newPassword string) error {
	// Get the user by ID
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update the user's password
	user.SetPasswordHash(string(hashedPassword))

	// Save the changes
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ForgotPassword generates a new password and sends it to the user's email
func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	// Get the user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// For security reasons, we don't reveal if the email exists or not
		// We return success even if the email doesn't exist
		return nil
	}

	// Generate a new random password
	newPassword := generateRandomPassword(12)

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update the user's password
	user.SetPasswordHash(string(hashedPassword))

	// Save the changes
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Send the new password by email
	subject := "Golene Evasion - Nouveau mot de passe"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Nouveau mot de passe - Golene Evasion</h2>
			<p>Cher/Chère %s %s,</p>
			<p>Votre demande de réinitialisation de mot de passe a été traitée.</p>
			<p>Voici votre nouveau mot de passe : <strong>%s</strong></p>
			<p>Nous vous recommandons de changer ce mot de passe après votre prochaine connexion.</p>
			<p>Cordialement,<br>L'équipe Golene Evasion</p>
		</body>
		</html>
	`, user.GetFirstName(), user.GetLastName(), newPassword)

	err = s.sendEmail(email, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send new password email: %w", err)
	}

	return nil
}

// GenerateRefereeInvitationToken creates a special JWT token for referee invitations
func (s *UserService) GenerateRefereeInvitationToken(ctx context.Context, competitionID int32) (string, int64, error) {
	expiresAt := time.Now().Add(time.Minute * 15).Unix()
	invitationClaims := jwt.MapClaims{
		"competition_id": competitionID,
		"type":           "referee_invitation",
		"iss":            "golene-evasion.com",
		"exp":            expiresAt,
	}

	invitationToken := jwt.NewWithClaims(jwt.SigningMethodHS256, invitationClaims)
	tokenString, err := invitationToken.SignedString([]byte(s.cfg.Jwt.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// Helper function to verify and extract competition ID from referee invitation token
func (s *UserService) verifyRefereeInvitationToken(token string) (int32, error) {
	// Parse the invitation token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.Jwt.SecretKey), nil
	})

	if err != nil || !parsedToken.Valid {
		return 0, ErrInvalidToken
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	// Verify token type
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "referee_invitation" {
		return 0, ErrInvalidToken
	}

	// Verify issuer
	if !claims.VerifyIssuer("golene-evasion.com", true) {
		return 0, ErrInvalidToken
	}

	// Extract competition ID
	var competitionID int32
	if id, ok := claims["competition_id"].(float64); ok {
		competitionID = int32(id)
	} else {
		return 0, ErrInvalidToken
	}

	return competitionID, nil
}

// AcceptRefereeInvitation processes a referee invitation token and adds the user to the competition
func (s *UserService) AcceptRefereeInvitation(ctx context.Context, token string, userEmail string) (*aggregate.JwtToken, error) {
	// Verify token and extract competition ID
	competitionID, err := s.verifyRefereeInvitationToken(token)
	if err != nil {
		return nil, err
	}

	// Get the user by email
	user, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Add referee role for the competition
	newRole := fmt.Sprintf("referee:%d", competitionID)
	user.AddRole(newRole)

	if len(user.GetRoles()) >= 500 {
		return nil, ErrMaximumRolesReached
	}

	// Save the changes
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to add referee role: %w", err)
	}

	// Generate new tokens for the user
	return s.generateTokens(user)
}

// AcceptRefereeInvitationUnauthenticated processes a referee invitation for unauthenticated users
func (s *UserService) AcceptRefereeInvitationUnauthenticated(ctx context.Context, token, firstName, lastName, email, password string) (*aggregate.JwtToken, error) {
	// Verify token and extract competition ID
	competitionID, err := s.verifyRefereeInvitationToken(token)
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// User exists, verify password
		if err := bcrypt.CompareHashAndPassword([]byte(existingUser.GetPasswordHash()), []byte(password)); err != nil {
			return nil, ErrInvalidCredentials
		}

		// Add referee role for the competition
		newRole := fmt.Sprintf("referee:%d", competitionID)
		existingUser.AddRole(newRole)

		if len(existingUser.GetRoles()) >= 500 {
			return nil, ErrMaximumRolesReached
		}

		// Save the changes
		err = s.userRepo.UpdateUser(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to add referee role: %w", err)
		}

		// Generate tokens for existing user
		return s.generateTokens(existingUser)
	}

	// User doesn't exist, create new user
	user := aggregate.NewUser()
	user.SetEmail(email)
	user.SetFirstName(firstName)
	user.SetLastName(lastName)

	// Set referee role for the specified competition
	role := fmt.Sprintf("referee:%d", competitionID)
	user.SetRoles(role)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.SetPasswordHash(string(hashedPassword))

	// Save the user to the database
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens for new user
	return s.generateTokens(user)
}
