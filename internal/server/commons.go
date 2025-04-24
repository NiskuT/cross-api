package server

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
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
