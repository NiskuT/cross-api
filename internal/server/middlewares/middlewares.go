package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/NiskuT/cross-api/internal/domain/entity"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func Authentication(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Retrieve the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		// Expect header in the format "Bearer <token>".
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenStr := parts[1]

		// Parse the token.
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract the claims as a map.
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		if claims.VerifyIssuer("golene-evasion.com", true) == false {
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
