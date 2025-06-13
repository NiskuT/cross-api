package server

import (
	"errors"
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
// @Param        Cookie  header string    true  "Authentication cookie"
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
	err = checkHasAccessToCompetition(c, int32(competitionID))
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
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
		Gender:        participant.GetGender(),
		Club:          participant.GetClub(),
	}

	c.JSON(http.StatusOK, response)
}

// createRun godoc
// @Summary      Create a new run
// @Description  Creates a new run and updates the liveranking
// @Tags         run
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
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
	err := checkHasAccessToCompetition(c, runInput.CompetitionID)
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
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

	run.SetRefereeId(user.Id)

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

// getParticipantRuns godoc
// @Summary      Get all runs for a participant
// @Description  Retrieves all runs for a specific participant with referee and zone information (admin only)
// @Tags         run
// @Accept       json
// @Produce      json
// @Param        Cookie        header    string  true   "Authentication cookie"
// @Param        competitionID path      int     true   "Competition ID"
// @Param        dossard       path      int     true   "Participant dossard number"
// @Success      200           {object}  models.RunListResponse     "Returns list of runs with details"
// @Failure      400           {object}  models.ErrorResponse       "Bad Request"
// @Failure      401           {object}  models.ErrorResponse       "Unauthorized"
// @Failure      403           {object}  models.ErrorResponse       "Forbidden (admin access required)"
// @Failure      404           {object}  models.ErrorResponse       "Not Found"
// @Failure      500           {object}  models.ErrorResponse       "Internal Server Error"
// @Router       /competition/{competitionID}/participant/{dossard}/runs [get]
func (s *Server) getParticipantRuns(c *gin.Context) {
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

	// Check if user has admin access to the competition
	err = checkHasAdminAccessToCompetition(c, int32(competitionID))
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
		return
	}

	// Get runs with details from service
	runs, err := s.runService.ListRunsByDossardWithDetails(c, int32(competitionID), int32(dossard))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Build response
	response := models.RunListResponse{
		Runs: make([]*models.RunDetailsResponse, 0, len(runs)),
	}

	for _, run := range runs {
		response.Runs = append(response.Runs, &models.RunDetailsResponse{
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
			RefereeID:     run.GetRefereeId(),
			RefereeName:   run.GetRefereeName(),
		})
	}

	c.JSON(http.StatusOK, response)
}

// updateRun godoc
// @Summary      Update a run
// @Description  Updates an existing run and recalculates liveranking (admin only)
// @Tags         run
// @Accept       json
// @Produce      json
// @Param        Cookie  header    string               true  "Authentication cookie"
// @Param        run     body      models.RunUpdateInput true  "Run update data"
// @Success      200     {object}  models.RunResponse   "Returns updated run data"
// @Failure      400     {object}  models.ErrorResponse "Bad Request"
// @Failure      401     {object}  models.ErrorResponse "Unauthorized"
// @Failure      403     {object}  models.ErrorResponse "Forbidden (admin access required)"
// @Failure      404     {object}  models.ErrorResponse "Run not found"
// @Failure      500     {object}  models.ErrorResponse "Internal Server Error"
// @Router       /run [put]
func (s *Server) updateRun(c *gin.Context) {
	var runInput models.RunUpdateInput
	if err := c.ShouldBindJSON(&runInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	// Check if user has admin access to the competition
	err := checkHasAdminAccessToCompetition(c, runInput.CompetitionID)
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
		return
	}

	// First verify the run exists
	existingRun, err := s.runService.GetRun(c, runInput.CompetitionID, runInput.RunNumber, runInput.Dossard)
	if err != nil {
		if errors.Is(err, repository.ErrRunNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("run not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Update the run with new values
	existingRun.SetZone(runInput.Zone)
	existingRun.SetDoor1(runInput.Door1)
	existingRun.SetDoor2(runInput.Door2)
	existingRun.SetDoor3(runInput.Door3)
	existingRun.SetDoor4(runInput.Door4)
	existingRun.SetDoor5(runInput.Door5)
	existingRun.SetDoor6(runInput.Door6)
	existingRun.SetPenality(runInput.Penality)
	existingRun.SetChronoSec(runInput.ChronoSec)

	// Update the run
	err = s.runService.UpdateRun(c, existingRun)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Build response
	response := models.RunResponse{
		CompetitionID: existingRun.GetCompetitionID(),
		Dossard:       existingRun.GetDossard(),
		RunNumber:     existingRun.GetRunNumber(),
		Zone:          existingRun.GetZone(),
		Door1:         existingRun.GetDoor1(),
		Door2:         existingRun.GetDoor2(),
		Door3:         existingRun.GetDoor3(),
		Door4:         existingRun.GetDoor4(),
		Door5:         existingRun.GetDoor5(),
		Door6:         existingRun.GetDoor6(),
		Penality:      existingRun.GetPenality(),
		ChronoSec:     existingRun.GetChronoSec(),
	}

	c.JSON(http.StatusOK, response)
}

// deleteRun godoc
// @Summary      Delete a run
// @Description  Deletes an existing run and recalculates liveranking (admin only)
// @Tags         run
// @Accept       json
// @Produce      json
// @Param        Cookie        header    string  true  "Authentication cookie"
// @Param        competitionID query     int     true  "Competition ID"
// @Param        dossard       query     int     true  "Participant dossard number"
// @Param        runNumber     query     int     true  "Run number"
// @Success      200           {object}  gin.H   "Run deleted successfully"
// @Failure      400           {object}  models.ErrorResponse "Bad Request"
// @Failure      401           {object}  models.ErrorResponse "Unauthorized"
// @Failure      403           {object}  models.ErrorResponse "Forbidden (admin access required)"
// @Failure      404           {object}  models.ErrorResponse "Run not found"
// @Failure      500           {object}  models.ErrorResponse "Internal Server Error"
// @Router       /run [delete]
func (s *Server) deleteRun(c *gin.Context) {
	competitionIDStr := c.Query("competitionID")
	dossardStr := c.Query("dossard")
	runNumberStr := c.Query("runNumber")

	if competitionIDStr == "" || dossardStr == "" || runNumberStr == "" {
		RespondError(c, http.StatusBadRequest, errors.New("competitionID, dossard, and runNumber are required"))
		return
	}

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

	runNumber, err := strconv.ParseInt(runNumberStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid run number"))
		return
	}

	// Check if user has admin access to the competition
	err = checkHasAdminAccessToCompetition(c, int32(competitionID))
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
		return
	}

	// Delete the run
	err = s.runService.DeleteRun(c, int32(competitionID), int32(runNumber), int32(dossard))
	if err != nil {
		if errors.Is(err, repository.ErrRunNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("run not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Run deleted successfully",
	})
}
