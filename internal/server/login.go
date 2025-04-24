package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/NiskuT/cross-api/internal/domain/models"
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
// @Success      200           {object}  userservice.User     				 "Returns user information and tokens in cookies"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /login [post]
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

	c.SetCookie(AccessToken, user.GetAccessToken(), 0, "/", "", true, true)
	c.SetCookie(RefreshToken, user.GetRefreshToken(), 0, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{})
}

// refresh godoc
// @Summary      Refresh a JWT token
// @Description  Takes a refresh token and returns a new access token (and possibly a new refresh token).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <refresh_token>"
// @Success      200            {object}  userservice.JwtToken   "New access/refresh token pair in http-only cookies"
// @Failure      400            {object}  models.ErrorResponse   "Bad Request"
// @Failure      401            {object}  models.ErrorResponse   "Unauthorized (missing or invalid token)"
// @Failure      500            {object}  models.ErrorResponse   "Internal Server Error"
// @Router       /refresh [get]
func (s *Server) refresh(c *gin.Context) {
	refreshRequest := c.Request.Header.Get("Authorization")
	if refreshRequest == "" {
		RespondError(c, http.StatusBadRequest, ErrMissingAuthorizationHeader)
		return
	}

	// Remove the 'Bearer ' prefix from the token
	// Expect header in the format "Bearer <token>".
	parts := strings.Split(refreshRequest, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
		return
	}
	tokenStr := parts[1]

	token, err := s.userService.RefreshToken(c, tokenStr)
	if err != nil {
		// If the refresh token is invalid or the user is not found, return Unauthorized.
		if err.Error() == "invalid refresh token" || err.Error() == "user not found" {
			RespondError(c, http.StatusUnauthorized, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.SetCookie(AccessToken, token.GetAccessToken(), 0, "/", "", true, true)
	c.SetCookie(RefreshToken, token.GetRefreshToken(), 0, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{})
}
