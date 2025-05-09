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

func checkHasAccessToCompetition(c *gin.Context, competitionID int32) error {
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", competitionID)) ||
		middlewares.HasRole(c, fmt.Sprintf("referee:%d", competitionID)) ||
		middlewares.HasRole(c, "admin:*")
	if !hasRole {
		return ErrForbidden
	}

	return nil
}
