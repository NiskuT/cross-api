package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/repository"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	serviceErr "github.com/NiskuT/cross-api/internal/service"
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

// createRun godoc
// @Summary      Create a new run
// @Description  Creates a new run and updates the liveranking
// @Tags         run
// @Accept       json
// @Produce      json
// @Param        run  body       models.RunInput  true  "Run data"
// @Success      201  {object}   models.RunResponse     "Returns created run data"
// @Failure      400  {object}   models.ErrorResponse   "Bad Request"
// @Failure      401  {object}   models.ErrorResponse   "Unauthorized"
// @Failure      404  {object}   models.ErrorResponse   "Not Found"
// @Failure      500  {object}   models.ErrorResponse   "Internal Server Error"
// @Router       /run [post]
func (s *Server) createRun(c *gin.Context) {
	var runInput models.RunInput
	if err := c.ShouldBindJSON(&runInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	// Check if user has appropriate role (admin or referee for the competition)
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", runInput.CompetitionID)) ||
		middlewares.HasRole(c, fmt.Sprintf("referee:%d", runInput.CompetitionID))
	if !hasRole {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Get referee ID from token
	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Create run aggregate
	run := aggregate.NewRun()
	run.SetCompetitionID(runInput.CompetitionID)
	run.SetDossard(runInput.Dossard)
	run.SetZone(runInput.Zone)
	run.SetDoor1(runInput.Door1)
	run.SetDoor2(runInput.Door2)
	run.SetDoor3(runInput.Door3)
	run.SetDoor4(runInput.Door4)
	run.SetDoor5(runInput.Door5)
	run.SetDoor6(runInput.Door6)
	run.SetPenality(runInput.Penality)
	run.SetChronoSec(runInput.ChronoSec)

	// Parse referee ID from user token ID (sub claim)
	refereeID, err := strconv.ParseInt(user.Id, 10, 32)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, errors.New("invalid referee ID"))
		return
	}
	run.SetRefereeId(int32(refereeID))

	// Call service to create run
	err = s.runService.CreateRun(c, run)
	if err != nil {
		// Determine appropriate error code based on error type
		if errors.Is(err, serviceErr.ErrInvalidRunData) {
			RespondError(c, http.StatusBadRequest, err)
		} else if errors.Is(err, repository.ErrParticipantNotFound) ||
			errors.Is(err, serviceErr.ErrScaleNotFound) {
			RespondError(c, http.StatusNotFound, err)
		} else {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}

	// Build response
	response := models.RunResponse{
		CompetitionID: run.GetCompetitionID(),
		Dossard:       run.GetDossard(),
		RunNumber:     run.GetRunNumber(),
		Zone:          run.GetZone(),
		Door1:         run.GetDoor1(),
		Door2:         run.GetDoor2(),
		Door3:         run.GetDoor3(),
		Door4:         run.GetDoor4(),
		Door5:         run.GetDoor5(),
		Door6:         run.GetDoor6(),
		Penality:      run.GetPenality(),
		ChronoSec:     run.GetChronoSec(),
	}

	c.JSON(http.StatusCreated, response)
}
