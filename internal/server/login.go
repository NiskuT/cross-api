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
