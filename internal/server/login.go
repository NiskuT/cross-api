package server

import (
	"errors"
	"net/http"

	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

// ErrMissingAuthorizationHeader indicates an Authorization header was not provided.
var (
	ErrMissingAuthorizationHeader = errors.New("missing authorization header")
	ErrInvalidPage                = errors.New("invalid page parameter")
	ErrInvalidLimit               = errors.New("invalid limit parameter")
)

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

// login godoc
// @Summary      Log in a user
// @Description  Authenticates a user with email and password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      models.LoginUser              true  "Login credentials"
// @Success      200           {object}  models.RoleResponse           "Returns user information and tokens in cookies"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /login [put]
func (s *Server) login(c *gin.Context) {
	var loginRequest models.LoginUser
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := s.userService.Login(c, loginRequest.Email, loginRequest.Password)
	if err != nil {
		// If the error indicates that the credentials are invalid or the user does not exist.
		if err.Error() == "user not found" || err.Error() == "invalid credentials" {
			RespondError(c, http.StatusUnauthorized, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	// Reset rate limit on successful login
	clientIP := s.rateLimiter.GetClientIP(c)
	s.rateLimiter.ResetAttempts("login", clientIP)

	c.SetCookie(AccessToken, user.GetAccessToken(), 0, "/", "", middlewares.SecureMode, true)
	c.SetCookie(RefreshToken, user.GetRefreshToken(), 0, "/", "", middlewares.SecureMode, true)

	c.JSON(http.StatusOK, models.RoleResponse{
		Roles: user.GetRoles(),
	})
}

// logout godoc
// @Summary      Log out a user
// @Description  Clears authentication cookies to log out the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200           {object}  gin.H                         "Successfully logged out"
// @Router       /logout [post]
func (s *Server) logout(c *gin.Context) {
	// Clear the access token cookie
	c.SetCookie(AccessToken, "", -1, "/", "", middlewares.SecureMode, true)

	// Clear the refresh token cookie
	c.SetCookie(RefreshToken, "", -1, "/", "", middlewares.SecureMode, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// changePassword godoc
// @Summary      Change user password
// @Description  Allows authenticated users to change their password by providing current and new password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Cookie              header    string                         true  "Authentication cookie"
// @Param        changePasswordRequest body      models.ChangePasswordInput     true  "Password change data"
// @Success      200                 {object}  gin.H                          "Password changed successfully"
// @Failure      400                 {object}  models.ErrorResponse           "Bad Request"
// @Failure      401                 {object}  models.ErrorResponse           "Unauthorized (invalid current password)"
// @Failure      500                 {object}  models.ErrorResponse           "Internal Server Error"
// @Router       /auth/password [put]
func (s *Server) changePassword(c *gin.Context) {
	var changePasswordRequest models.ChangePasswordInput
	if err := c.ShouldBindJSON(&changePasswordRequest); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	// Get user from context (set by authentication middleware)
	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	// Call service to change password
	err = s.userService.ChangePassword(
		c.Request.Context(),
		user.Id,
		changePasswordRequest.CurrentPassword,
		changePasswordRequest.NewPassword,
	)

	if err != nil {
		if err.Error() == "invalid email or password" {
			RespondError(c, http.StatusUnauthorized, errors.New("current password is incorrect"))
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// forgotPassword godoc
// @Summary      Reset forgotten password
// @Description  Generates a new password and sends it to the user's email address
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        forgotPasswordRequest body      models.ForgotPasswordInput     true  "Email for password reset"
// @Success      200                   {object}  gin.H                          "Password reset email sent"
// @Failure      400                   {object}  models.ErrorResponse           "Bad Request"
// @Failure      500                   {object}  models.ErrorResponse           "Internal Server Error"
// @Router       /auth/forgot-password [post]
func (s *Server) forgotPassword(c *gin.Context) {
	var forgotPasswordRequest models.ForgotPasswordInput
	if err := c.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	// Call service to reset password
	err := s.userService.ForgotPassword(c.Request.Context(), forgotPasswordRequest.Email)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Always return success for security reasons (don't reveal if email exists)
	c.JSON(http.StatusOK, gin.H{
		"message": "If the email address exists in our system, a new password has been sent to it",
	})
}
