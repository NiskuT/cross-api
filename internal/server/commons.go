package server

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("the user is not authorized to access this resource")
)

func getPagination(c *gin.Context) (int32, int32) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// Support both 'page_size' and 'limit' for backward compatibility
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("limit", "10")))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	return int32(page), int32(pageSize)
}

func checkHasAccessToCompetition(c *gin.Context, competitionID int32) error {
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", competitionID)) ||
		middlewares.HasRole(c, fmt.Sprintf("referee:%d", competitionID)) ||
		middlewares.HasRole(c, "admin:*")
	if !hasRole {
		return ErrForbidden
	}

	return nil
}

// checkHasAdminAccessToCompetition checks if user is admin of the competition or super admin
// This is stricter than checkHasAccessToCompetition as it excludes regular referees
func checkHasAdminAccessToCompetition(c *gin.Context, competitionID int32) error {
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", competitionID)) ||
		middlewares.HasRole(c, "admin:*")
	if !hasRole {
		return ErrForbidden
	}

	return nil
}
