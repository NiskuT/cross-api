package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NiskuT/cross-api/internal/domain/entity"
	"github.com/NiskuT/cross-api/internal/domain/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

// extractAccessToken retrieves the access token from the cookie
func extractAccessToken(c *gin.Context) (string, error) {
	tokenStr, err := c.Cookie(AccessToken)
	if err != nil {
		return "", errors.New("authorization cookie missing")
	}

	if tokenStr == "" {
		return "", errors.New("empty token")
	}

	return tokenStr, nil
}

// parseAndValidateToken parses the JWT token and validates it
func parseAndValidateToken(tokenStr, secretKey string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
}

// handleExpiredToken processes refresh token logic when the access token has expired
func handleExpiredToken(c *gin.Context, userService service.UserService) (bool, error) {
	refreshToken, err := c.Cookie(RefreshToken)
	if err != nil || refreshToken == "" {
		return false, errors.New("refresh token missing")
	}

	tokens, err := userService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		return false, errors.New("invalid refresh token")
	}

	c.SetCookie(AccessToken, tokens.GetAccessToken(), 0, "/", "", true, true)
	c.SetCookie(RefreshToken, tokens.GetRefreshToken(), 0, "/", "", true, true)

	return true, nil
}

// extractUserFromClaims builds a UserToken from JWT claims
func extractUserFromClaims(claims jwt.MapClaims) (entity.UserToken, error) {
	var customClaims entity.UserToken

	// Verify issuer
	if !claims.VerifyIssuer("golene-evasion.com", true) {
		return customClaims, errors.New("invalid token issuer")
	}

	// Extract user ID and email
	if sub, ok := claims["sub"].(string); ok {
		customClaims.Id = sub
	}
	if email, ok := claims["email"].(string); ok {
		customClaims.Email = email
	}

	// Extract roles
	if rolesVal, ok := claims["roles"]; ok {
		switch roles := rolesVal.(type) {
		case []interface{}:
			for _, r := range roles {
				if roleStr, ok := r.(string); ok {
					customClaims.Roles = append(customClaims.Roles, roleStr)
				}
			}
		case []string:
			customClaims.Roles = roles
		}
	}

	return customClaims, nil
}

func Authentication(secretKey string, userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		var err error
		var refreshed bool

	tokenValidation:
		// Step 1: Extract access token
		tokenStr, err = extractAccessToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Step 2: Parse and validate token
		token, err := parseAndValidateToken(tokenStr, secretKey)
		if err != nil {
			// Check specifically for token expiration
			if !refreshed {
				validationErr, ok := err.(*jwt.ValidationError)
				if ok && validationErr.Errors&jwt.ValidationErrorExpired != 0 {
					// Step 3: Handle expired token with refresh flow
					refreshSuccessful, refreshErr := handleExpiredToken(c, userService)
					if refreshSuccessful {
						refreshed = true
						goto tokenValidation // Restart validation with new token
					}
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": refreshErr.Error()})
					return
				}
			}

			// Generic error for other token validation failures
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Verify token is valid
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Step 4: Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Step 5: Extract user from claims
		customClaims, err := extractUserFromClaims(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Step 6: Attach user to context
		c.Set("user", customClaims)
		c.Next()
	}
}

func GetUser(c *gin.Context) (*entity.UserToken, error) {
	val, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}

	userClaims, ok := val.(entity.UserToken)
	if !ok {
		return nil, errors.New("user claims have an unexpected type")
	}

	return &userClaims, nil
}

func HasRole(c *gin.Context, role string) bool {
	user, err := GetUser(c)
	if err != nil {
		return false
	}

	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}
