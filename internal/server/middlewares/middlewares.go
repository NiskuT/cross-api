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

func Authentication(secretKey string, userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Retrieve token from cookie
		tokenStr, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie missing"})
			return
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Empty token"})
			return
		}

		// Parse the token.
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		// Handle specific token errors
		if err != nil {
			// Check specifically for token expiration
			if validationErr, ok := err.(*jwt.ValidationError); ok {
				if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
					refreshToken, err := c.Cookie("refresh_token")
					if err != nil {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
						return
					}

					if refreshToken == "" {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
						return
					}

					tokens, err := userService.RefreshToken(c.Request.Context(), refreshToken)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
						return
					}

					c.SetCookie(AccessToken, tokens.GetAccessToken(), 0, "/", "", true, true)
					c.SetCookie(RefreshToken, tokens.GetRefreshToken(), 0, "/", "", true, true)

					c.Next()
					return
				}
			}

			// Generic error for other token validation failures
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract the claims as a map.
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		if !claims.VerifyIssuer("golene-evasion.com", true) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token issuer"})
			return
		}

		// Fill our custom claims struct.
		var customClaims entity.UserToken
		if sub, ok := claims["sub"].(string); ok {
			customClaims.Id = sub
		}
		if email, ok := claims["email"].(string); ok {
			customClaims.Email = email
		}

		// For roles, the type may be []interface{} so we need to convert it.
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

		// Attach the custom claims to the context.
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
