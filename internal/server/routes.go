package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/orkys/backend/gateway/internal/domain/models"
	"gitlab.com/orkys/backend/gateway/internal/server/middlewares"
	"gitlab.com/orkys/backend/gateway/internal/utils"
	userservice "gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/user-service"
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

	loginRequestProto := userservice.LoginRequest{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	}

	user, err := s.userService.LoginUser(c, &loginRequestProto)
	if err != nil {
		// If the error indicates that the credentials are invalid or the user does not exist.
		if err.Error() == "user not found" || err.Error() == "invalid credentials" {
			RespondError(c, http.StatusUnauthorized, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	log.Info().Msgf("User %s logged in", user.Token)

	c.SetCookie(AccessToken, user.Token, 0, "/", "", true, true)
	c.SetCookie(RefreshToken, user.RefreshToken, 0, "/", "", true, true)

	c.JSON(http.StatusOK, user.User)
}

// register godoc
// @Summary      Register a new user
// @Description  Creates a new user in the system and returns the user object with tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body   models.CreateUser       true  "Registration details"
// @Success      201           {object}  userservice.User				       "Created user with tokens in cookies"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      409           {object}  models.ErrorResponse          "Conflict (user already exists)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /register [post]
func (s *Server) register(c *gin.Context) {
	var registerRequest models.CreateUser
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	registerRequestProto := userservice.User{
		Email:     registerRequest.Email,
		Password:  registerRequest.Password,
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
	}

	user, err := s.userService.RegisterUser(c, &registerRequestProto)
	if err != nil {
		if err.Error() == "user already exists" {
			RespondError(c, http.StatusConflict, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.SetCookie(AccessToken, user.Token, 0, "/", "", true, true)
	c.SetCookie(RefreshToken, user.RefreshToken, 0, "/", "", true, true)

	c.JSON(http.StatusCreated, user.User)
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

	refreshRequestProto := userservice.JwtToken{RefreshToken: tokenStr}
	token, err := s.userService.RefreshToken(c, &refreshRequestProto)
	if err != nil {
		// If the refresh token is invalid or the user is not found, return Unauthorized.
		if err.Error() == "invalid refresh token" || err.Error() == "user not found" {
			RespondError(c, http.StatusUnauthorized, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.SetCookie(AccessToken, token.Token, 0, "/", "", true, true)
	c.SetCookie(RefreshToken, token.RefreshToken, 0, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{})
}

// list godoc
// @Summary      List users
// @Description  Returns a list of users, limited by a query param (default 10).
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true               "Bearer <token>"
// @Param        page           query     int     false 						 "Page number for pagination"
// @Param        limit          query     int     false 						 "Number of events per page"
// @Success      200            {object}  userservice.UsersResponse  "List of users"
// @Failure      400            {object}  models.ErrorResponse       "Bad Request"
// @Failure      401            {object}  models.ErrorResponse       "Unauthorized"
// @Failure      500            {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /list [get]
func (s *Server) list(c *gin.Context) {

	_, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	page, limit := getPagination(c)

	numberProto := userservice.ListUsersRequest{PageNumber: int32(page), PageSize: int32(limit)}
	users, err := s.userService.ListUsers(c, &numberProto)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// oauthInitiate godoc
// @Summary      Initiate an OAuth flow for a provider
// @Description  Initiates an OAuth flow (for providers "custom" or "google") by calling the user service and returning the provider’s auth URL.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        provider  path      string  true  "OAuth provider (custom or google)"
// @Param        state     query     string  false "Optional state parameter. If omitted, a new state will be generated."
// @Success      200       {object}  userservice.OAuthInitiateResponse "OAuth initiation response"
// @Failure      400       {object}  models.ErrorResponse               "Bad Request"
// @Failure      500       {object}  models.ErrorResponse               "Internal Server Error"
// @Router       /oauth/{provider}/initiate [get]
func (s *Server) oauthInitiate(c *gin.Context) {
	provider := c.Param("provider")
	// Validate provider – here we support "custom" and "google"
	if provider != "google" {
		RespondError(c, http.StatusBadRequest, errors.New("unsupported oauth provider"))
		return
	}

	// Retrieve an optional state from query parameters or generate a new one.
	state := c.Query("state")
	if state == "" {
		state = utils.GenerateState()
	}

	req := &userservice.OAuthInitiateRequest{
		Provider: provider,
		State:    state,
	}

	resp, err := s.authService.InitiateOAuth(c, req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Return the auth URL in JSON. The client can then redirect the user to this URL.
	c.JSON(http.StatusOK, resp)
}

// oauthCallback godoc
// @Summary      Process the OAuth callback from a provider
// @Description  Processes the OAuth callback from the provider, calls the user service, and returns a login response with JWT tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        provider  path      string  true  									"OAuth provider (custom or google)"
// @Param        state     query     string  true  									"State parameter for CSRF validation"
// @Param        code      query     string  false 									"Authorization code (and any other provider parameters)"
// @Success      200       {object}  userservice.User 							"User login response with tokens in cookies"
// @Failure      400       {object}  models.ErrorResponse          	"Bad Request"
// @Failure      500       {object}  models.ErrorResponse          	"Internal Server Error"
// @Router       /oauth/{provider}/callback [get]
func (s *Server) oauthCallback(c *gin.Context) {
	provider := c.Param("provider")
	// Validate provider – here we support "custom" and "google"
	if provider != "google" {
		RespondError(c, http.StatusBadRequest, errors.New("unsupported oauth provider"))
		return
	}

	state := c.Query("state")
	if state == "" {
		RespondError(c, http.StatusBadRequest, errors.New("state parameter is missing"))
		return
	}

	// Gather all query parameters from the callback into a map[string]string.
	// c.Request.URL.Query() returns a map[string][]string, so we take the first value.
	q := c.Request.URL.Query()
	queryParams := make(map[string]string)
	for key, values := range q {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	req := &userservice.OAuthCallbackRequest{
		Provider:    provider,
		State:       state,
		QueryParams: queryParams,
	}

	resp, err := s.authService.OAuthCallback(c, req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.SetCookie(AccessToken, resp.Token, 0, "/", "", true, true)
	c.SetCookie(RefreshToken, resp.RefreshToken, 0, "/", "", true, true)

	// Return the login response (user info and tokens) to the frontend.
	c.JSON(http.StatusOK, resp.User)
}

// register godoc
// @Summary      Update user details
// @Description  Updates the user details (first name, last name, password and avatar URL) if they are not empty.
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  								true   	"Bearer <token>"
// @Param        UpdateUserRequest  body models.UpdateUser      true   	"User details to update"
// @Success      200           {object}  commons.Response			     			"Created user with tokens"
// @Failure      400           {object}  models.ErrorResponse          	"Bad Request"
// @Failure      401           {object}  models.ErrorResponse   				"Unauthorized (missing or invalid token)"
// @Failure      500           {object}  models.ErrorResponse          	"Internal Server Error"
// @Router       /user [patch]
func (s *Server) updateUser(c *gin.Context) {
	var updateUserRequest models.UpdateUser
	if err := c.ShouldBindJSON(&updateUserRequest); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err)
		return
	}

	updateUserRequestProto := userservice.UpdateUserRequest{
		Id:        user.Id,
		FirstName: updateUserRequest.FirstName,
		LastName:  updateUserRequest.LastName,
		Password:  updateUserRequest.Password,
		// Roles:       user.Roles,
		AvatarUrl: updateUserRequest.AvatarUrl,
	}

	resp, err := s.userService.UpdateUser(c, &updateUserRequestProto)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func getPagination(c *gin.Context) (int32, int32) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		page = 0
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10")) // Default limit to 10
	if err != nil {
		limit = 10
	}

	return int32(page), int32(limit)
}
