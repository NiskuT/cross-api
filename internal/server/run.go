package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

// getParticipant godoc
// @Summary      Get participant information
// @Description  Retrieves a participant's information based on dossard number and competition ID
// @Tags         participant
// @Accept       json
// @Produce      json
// @Param        competitionID  path      int     true  "Competition ID"
// @Param        dossard        path      int     true  "Dossard Number"
// @Success      200            {object}  models.ParticipantResponse  "Returns participant data"
// @Failure      400            {object}  models.ErrorResponse        "Bad Request"
// @Failure      401            {object}  models.ErrorResponse        "Unauthorized (invalid credentials)"
// @Failure      404            {object}  models.ErrorResponse        "Participant not found"
// @Failure      500            {object}  models.ErrorResponse        "Internal Server Error"
// @Router       /competition/{competitionID}/participant/{dossard} [get]
func (s *Server) getParticipant(c *gin.Context) {
	competitionIDStr := c.Param("competitionID")
	dossardStr := c.Param("dossard")

	competitionID, err := strconv.ParseInt(competitionIDStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid competition ID"))
		return
	}

	dossard, err := strconv.ParseInt(dossardStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid dossard number"))
		return
	}

	// Check if user has access to the competition
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", competitionID)) ||
		middlewares.HasRole(c, fmt.Sprintf("referee:%d", competitionID))
	if !hasRole {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Get participant through the service
	participant, err := s.competitionService.GetParticipant(c, int32(competitionID), int32(dossard))
	if err != nil {
		RespondError(c, http.StatusNotFound, errors.New("participant not found"))
		return
	}

	// Build response
	response := models.ParticipantResponse{
		CompetitionID: participant.GetCompetitionID(),
		DossardNumber: participant.GetDossardNumber(),
		FirstName:     participant.GetFirstName(),
		LastName:      participant.GetLastName(),
		Category:      participant.GetCategory(),
	}

	c.JSON(http.StatusOK, response)
}
