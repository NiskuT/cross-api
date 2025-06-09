package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/repository"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

// createCompetition godoc
// @Summary      Create a competition
// @Description  Creates a new competition and returns a JWT token.
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competition  body       models.Competition              true  "Competition data"
// @Success      200           {object}  models.CompetitionResponse     			 "Returns competition data"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /competition [post]
func (s *Server) createCompetition(c *gin.Context) {
	var competition models.Competition
	if err := c.ShouldBindJSON(&competition); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	if !middlewares.HasRole(c, "create:competition") {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	competitionAggregate := aggregate.NewCompetition()
	competitionAggregate.SetName(competition.Name)
	competitionAggregate.SetDescription(competition.Description)
	competitionAggregate.SetDate(competition.Date)
	competitionAggregate.SetLocation(competition.Location)
	competitionAggregate.SetOrganizer(competition.Organizer)
	competitionAggregate.SetContact(competition.Contact)

	competitionID, err := s.competitionService.CreateCompetition(c, competitionAggregate)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	user, err := middlewares.GetUser(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	newToken, err := s.userService.SetUserAsAdmin(c, user.Email, competitionID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.SetCookie(AccessToken, newToken.GetAccessToken(), 0, "/", "", middlewares.SecureMode, true)
	c.SetCookie(RefreshToken, newToken.GetRefreshToken(), 0, "/", "", middlewares.SecureMode, true)

	res := models.CompetitionResponse{
		ID:          competitionID,
		Name:        competition.Name,
		Description: competition.Description,
		Date:        competition.Date,
		Location:    competition.Location,
		Organizer:   competition.Organizer,
		Contact:     competition.Contact,
	}

	c.JSON(http.StatusOK, res)
}

// listCompetitions godoc
// @Summary      List competitions
// @Description  Lists all competitions
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Success      200           {object}  models.CompetitionListResponse     			 "Returns competition data"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /competition [get]
func (s *Server) listCompetitions(c *gin.Context) {
	competitions, err := s.competitionService.ListCompetitions(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	res := models.CompetitionListResponse{
		Competitions: make([]*models.CompetitionResponse, len(competitions)),
	}
	for i, competition := range competitions {
		res.Competitions[i] = &models.CompetitionResponse{
			ID:          competition.GetID(),
			Name:        competition.GetName(),
			Description: competition.GetDescription(),
			Date:        competition.GetDate(),
			Location:    competition.GetLocation(),
			Organizer:   competition.GetOrganizer(),
			Contact:     competition.GetContact(),
		}
	}
	c.JSON(http.StatusOK, res)
}

// addZoneToCompetition godoc
// @Summary      Add a zone to a competition
// @Description  Adds a zone to a competition
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competition  body       models.CompetitionScaleInput  true  "Competition data"
// @Success      200           {object}  gin.H       			 						 "Returns competition data"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /competition/zone [post]
func (s *Server) addZoneToCompetition(c *gin.Context) {
	var competitionScaleInput models.CompetitionScaleInput
	if err := c.ShouldBindJSON(&competitionScaleInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	if competitionScaleInput.CompetitionID == 0 {
		RespondError(c, http.StatusBadRequest, errors.New("competition ID is required"))
		return
	}

	role := fmt.Sprintf("admin:%d", competitionScaleInput.CompetitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	scale := aggregate.NewScale()
	scale.SetCompetitionID(competitionScaleInput.CompetitionID)
	scale.SetCategory(competitionScaleInput.Category)
	scale.SetZone(competitionScaleInput.Zone)
	scale.SetPointsDoor1(competitionScaleInput.PointsDoor1)
	scale.SetPointsDoor2(competitionScaleInput.PointsDoor2)
	scale.SetPointsDoor3(competitionScaleInput.PointsDoor3)
	scale.SetPointsDoor4(competitionScaleInput.PointsDoor4)
	scale.SetPointsDoor5(competitionScaleInput.PointsDoor5)
	scale.SetPointsDoor6(competitionScaleInput.PointsDoor6)

	err := s.competitionService.AddScale(c, competitionScaleInput.CompetitionID, scale)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone added to competition"})
}

// addParticipantsToCompetition godoc
// @Summary      Add participants to a competition
// @Description  Adds multiple participants to a competition from a CSV or Excel file
// @Tags         competition
// @Accept       multipart/form-data
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competitionID  formData  int     true  "Competition ID"
// @Param        file           formData  file    true  "CSV or Excel file with participants data (format: dossard number, category, last name, first name, gender)"
// @Success      200           {object}  gin.H                        "Successfully added participants"
// @Failure      400           {object}  models.ErrorResponse         "Bad Request"
// @Failure      401           {object}  models.ErrorResponse         "Unauthorized (invalid credentials)"
// @Failure      500           {object}  models.ErrorResponse         "Internal Server Error"
// @Router       /competition/participants [post]
func (s *Server) addParticipantsToCompetition(c *gin.Context) {
	competitionIDStr := c.PostForm("competitionID")
	if competitionIDStr == "" {
		RespondError(c, http.StatusBadRequest, errors.New("competition ID is required"))
		return
	}

	competitionIDInt, err := strconv.ParseInt(competitionIDStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid competition ID format"))
		return
	}
	competitionID := int32(competitionIDInt)

	role := fmt.Sprintf("admin:%d", competitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("file is required"))
		return
	}
	defer file.Close()

	// Get filename from the file header
	filename := fileHeader.Filename

	err = s.competitionService.AddParticipants(c, competitionID, file, filename)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participants added to competition"})
}

// addRefereeToCompetition godoc
// @Summary      Add a referee to a competition
// @Description  Invites a user as a referee to a competition
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        referee  body       models.RefereeInput  true  "Referee data"
// @Success      200      {object}   gin.H               "Successfully added referee"
// @Failure      400      {object}   models.ErrorResponse "Bad Request"
// @Failure      401      {object}   models.ErrorResponse "Unauthorized (invalid credentials)"
// @Failure      500      {object}   models.ErrorResponse "Internal Server Error"
// @Router       /competition/referee [post]
func (s *Server) addRefereeToCompetition(c *gin.Context) {
	var refereeInput models.RefereeInput
	if err := c.ShouldBindJSON(&refereeInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	if refereeInput.CompetitionID == 0 {
		RespondError(c, http.StatusBadRequest, errors.New("competition ID is required"))
		return
	}

	if refereeInput.FirstName == "" {
		RespondError(c, http.StatusBadRequest, errors.New("first name is required"))
		return
	}

	if refereeInput.LastName == "" {
		RespondError(c, http.StatusBadRequest, errors.New("last name is required"))
		return
	}

	if refereeInput.Email == "" {
		RespondError(c, http.StatusBadRequest, errors.New("email is required"))
		return
	}

	role := fmt.Sprintf("admin:%d", refereeInput.CompetitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	competition, err := s.competitionService.GetCompetition(c, refereeInput.CompetitionID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	err = s.userService.InviteUser(c, refereeInput.FirstName, refereeInput.LastName, refereeInput.Email, competition)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Referee added to competition"})
}

// listZones godoc
// @Summary      List zones for a competition
// @Description  Lists all available zones for a competition
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competitionID  path      int     true  "Competition ID"
// @Success      200           {object}  models.ZonesListResponse     "Returns list of zones"
// @Failure      400           {object}  models.ErrorResponse         "Bad Request"
// @Failure      401           {object}  models.ErrorResponse         "Unauthorized (invalid credentials)"
// @Failure      404           {object}  models.ErrorResponse         "Competition not found"
// @Failure      500           {object}  models.ErrorResponse         "Internal Server Error"
// @Router       /competition/{competitionID}/zones [get]
func (s *Server) listZones(c *gin.Context) {
	competitionIDStr := c.Param("competitionID")

	competitionID, err := strconv.ParseInt(competitionIDStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid competition ID"))
		return
	}

	// Check if user has access to the competition
	err = checkHasAccessToCompetition(c, int32(competitionID))
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
		return
	}

	// Get zones from service
	zones, err := s.competitionService.ListZones(c, int32(competitionID))
	if err != nil {
		if errors.Is(err, repository.ErrCompetitionNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("competition not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Build response
	response := models.ZonesListResponse{
		CompetitionID: int32(competitionID),
		Zones:         make([]models.ZoneResponse, 0, len(zones)),
	}

	for _, zone := range zones {
		scale, err := s.competitionService.GetScale(c, int32(competitionID), zone.GetCategory(), zone.GetZone())
		if err != nil {
			RespondError(c, http.StatusInternalServerError, err)
			return
		}
		response.Zones = append(response.Zones, models.ZoneResponse{
			Zone:        zone.GetZone(),
			Category:    zone.GetCategory(),
			PointsDoor1: scale.GetPointsDoor1(),
			PointsDoor2: scale.GetPointsDoor2(),
			PointsDoor3: scale.GetPointsDoor3(),
			PointsDoor4: scale.GetPointsDoor4(),
			PointsDoor5: scale.GetPointsDoor5(),
			PointsDoor6: scale.GetPointsDoor6(),
		})
	}

	c.JSON(http.StatusOK, response)
}

// updateZoneInCompetition godoc
// @Summary      Update a zone in a competition
// @Description  Updates an existing zone in a competition
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competition  body       models.CompetitionScaleInput  true  "Competition data"
// @Success      200           {object}  gin.H       			 						 "Returns success message"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      404           {object}  models.ErrorResponse          "Zone not found"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /competition/zone [put]
func (s *Server) updateZoneInCompetition(c *gin.Context) {
	var competitionScaleInput models.CompetitionScaleInput
	if err := c.ShouldBindJSON(&competitionScaleInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	if competitionScaleInput.CompetitionID == 0 {
		RespondError(c, http.StatusBadRequest, errors.New("competition ID is required"))
		return
	}

	role := fmt.Sprintf("admin:%d", competitionScaleInput.CompetitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	scale := aggregate.NewScale()
	scale.SetCompetitionID(competitionScaleInput.CompetitionID)
	scale.SetCategory(competitionScaleInput.Category)
	scale.SetZone(competitionScaleInput.Zone)
	scale.SetPointsDoor1(competitionScaleInput.PointsDoor1)
	scale.SetPointsDoor2(competitionScaleInput.PointsDoor2)
	scale.SetPointsDoor3(competitionScaleInput.PointsDoor3)
	scale.SetPointsDoor4(competitionScaleInput.PointsDoor4)
	scale.SetPointsDoor5(competitionScaleInput.PointsDoor5)
	scale.SetPointsDoor6(competitionScaleInput.PointsDoor6)

	err := s.competitionService.UpdateScale(c, competitionScaleInput.CompetitionID, scale)
	if err != nil {
		if errors.Is(err, repository.ErrScaleNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("zone not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone updated successfully"})
}

// deleteZoneFromCompetition godoc
// @Summary      Delete a zone from a competition
// @Description  Deletes an existing zone from a competition
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        zone  body       models.CompetitionZoneDeleteInput  true  "Zone deletion data"
// @Success      200           {object}  gin.H       			 						 "Returns success message"
// @Failure      400           {object}  models.ErrorResponse          "Bad Request"
// @Failure      401           {object}  models.ErrorResponse          "Unauthorized (invalid credentials)"
// @Failure      404           {object}  models.ErrorResponse          "Zone not found"
// @Failure      500           {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /competition/zone [delete]
func (s *Server) deleteZoneFromCompetition(c *gin.Context) {
	var zoneDeleteInput models.CompetitionZoneDeleteInput
	if err := c.ShouldBindJSON(&zoneDeleteInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	if zoneDeleteInput.CompetitionID == 0 {
		RespondError(c, http.StatusBadRequest, errors.New("competition ID is required"))
		return
	}

	if zoneDeleteInput.Category == "" {
		RespondError(c, http.StatusBadRequest, errors.New("category is required"))
		return
	}

	if zoneDeleteInput.Zone == "" {
		RespondError(c, http.StatusBadRequest, errors.New("zone is required"))
		return
	}

	role := fmt.Sprintf("admin:%d", zoneDeleteInput.CompetitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	err := s.competitionService.DeleteScale(c, zoneDeleteInput.CompetitionID, zoneDeleteInput.Category, zoneDeleteInput.Zone)
	if err != nil {
		if errors.Is(err, repository.ErrScaleNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("zone not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone deleted successfully"})
}

// getLiveranking godoc
// @Summary      Get live ranking
// @Description  Retrieves live ranking for a competition with optional category and gender filtering
// @Tags         competition
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competitionID  path      int     true  "Competition ID"
// @Param        category       query     string  false "Category filter (optional)"
// @Param        gender         query     string  false "Gender filter (optional, H or F)"
// @Param        page           query     int     false "Page number (default: 1)"
// @Param        page_size      query     int     false "Page size (default: 10)"
// @Success      200           {object}  models.LiverankingListResponse     "Returns live ranking data"
// @Failure      400           {object}  models.ErrorResponse               "Bad Request"
// @Failure      401           {object}  models.ErrorResponse               "Unauthorized (invalid credentials)"
// @Failure      404           {object}  models.ErrorResponse               "Competition not found"
// @Failure      500           {object}  models.ErrorResponse               "Internal Server Error"
// @Router       /competition/{competitionID}/liveranking [get]
func (s *Server) getLiveranking(c *gin.Context) {
	competitionIDStr := c.Param("competitionID")

	competitionID, err := strconv.ParseInt(competitionIDStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid competition ID"))
		return
	}

	// Check if user has access to the competition
	err = checkHasAccessToCompetition(c, int32(competitionID))
	if err != nil {
		RespondError(c, http.StatusForbidden, err)
		return
	}

	// Get query parameters
	category := c.Query("category")
	gender := c.Query("gender")
	page, pageSize := getPagination(c)

	// Validate gender parameter if provided
	if gender == "" || (gender != "H" && gender != "F") {
		RespondError(c, http.StatusBadRequest, errors.New("gender must be 'H' or 'F'"))
		return
	}

	// Get live ranking from service
	rankings, total, err := s.competitionService.GetLiveranking(c, int32(competitionID), category, gender, page, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrCompetitionNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("competition not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Build response
	response := models.LiverankingListResponse{
		CompetitionID: int32(competitionID),
		Category:      category,
		Gender:        gender,
		Page:          page,
		PageSize:      pageSize,
		Total:         total,
		Rankings:      make([]models.LiverankingResponse, 0, len(rankings)),
	}

	// Calculate rank based on position (considering pagination)
	baseRank := (page-1)*pageSize + 1

	for i, ranking := range rankings {
		response.Rankings = append(response.Rankings, models.LiverankingResponse{
			Rank:         baseRank + int32(i),
			Dossard:      ranking.GetDossard(),
			FirstName:    ranking.GetFirstName(),
			LastName:     ranking.GetLastName(),
			Category:     ranking.GetCategory(),
			Gender:       ranking.GetGender(),
			NumberOfRuns: ranking.GetNumberOfRuns(),
			TotalPoints:  ranking.GetTotalPoints(),
			Penality:     ranking.GetPenality(),
			ChronoSec:    ranking.GetChronoSec(),
		})
	}

	c.JSON(http.StatusOK, response)
}

// createParticipant godoc
// @Summary      Create a participant
// @Description  Creates a single participant for a competition
// @Tags         participant
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        participant  body       models.ParticipantInput  true  "Participant data"
// @Success      201           {object}  models.ParticipantResponse     "Returns created participant data"
// @Failure      400           {object}  models.ErrorResponse           "Bad Request"
// @Failure      401           {object}  models.ErrorResponse           "Unauthorized (invalid credentials)"
// @Failure      409           {object}  models.ErrorResponse           "Participant already exists"
// @Failure      500           {object}  models.ErrorResponse           "Internal Server Error"
// @Router       /participant [post]
func (s *Server) createParticipant(c *gin.Context) {
	var participantInput models.ParticipantInput
	if err := c.ShouldBindJSON(&participantInput); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	// Check if user has access to the competition
	role := fmt.Sprintf("admin:%d", participantInput.CompetitionID)
	if !middlewares.HasRole(c, role) {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Create participant aggregate
	participant := aggregate.NewParticipant()
	participant.SetCompetitionID(participantInput.CompetitionID)
	participant.SetDossardNumber(participantInput.DossardNumber)
	participant.SetFirstName(participantInput.FirstName)
	participant.SetLastName(participantInput.LastName)
	participant.SetCategory(participantInput.Category)
	participant.SetGender(participantInput.Gender)

	// Create participant through service
	err := s.competitionService.CreateParticipant(c, participant)
	if err != nil {
		// Check for duplicate participant error (need to check the error message since it's in different package)
		if errors.Is(err, repository.ErrCompetitionNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("competition not found"))
			return
		}
		// Check if it's a duplicate error from the participant repository
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			RespondError(c, http.StatusConflict, errors.New("participant with this dossard number already exists"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
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
	}

	c.JSON(http.StatusCreated, response)
}

// listParticipantsByCategory godoc
// @Summary      List participants by category
// @Description  Lists all participants for a competition filtered by category
// @Tags         participant
// @Accept       json
// @Produce      json
// @Param        Cookie  header string    true  "Authentication cookie"
// @Param        competitionID  path      int     true  "Competition ID"
// @Param        category       query     string  true  "Category filter"
// @Success      200           {object}  models.ParticipantListResponse "Returns list of participants"
// @Failure      400           {object}  models.ErrorResponse           "Bad Request"
// @Failure      401           {object}  models.ErrorResponse           "Unauthorized (invalid credentials)"
// @Failure      404           {object}  models.ErrorResponse           "Competition not found"
// @Failure      500           {object}  models.ErrorResponse           "Internal Server Error"
// @Router       /competition/{competitionID}/participants [get]
func (s *Server) listParticipantsByCategory(c *gin.Context) {
	competitionIDStr := c.Param("competitionID")
	category := c.Query("category")

	// Validate inputs
	competitionID, err := strconv.ParseInt(competitionIDStr, 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errors.New("invalid competition ID"))
		return
	}

	if category == "" {
		RespondError(c, http.StatusBadRequest, errors.New("category parameter is required"))
		return
	}

	// Check if user has access to the competition
	hasRole := middlewares.HasRole(c, fmt.Sprintf("admin:%d", competitionID)) ||
		middlewares.HasRole(c, fmt.Sprintf("referee:%d", competitionID))
	if !hasRole {
		RespondError(c, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Get participants through service
	participants, err := s.competitionService.ListParticipantsByCategory(c, int32(competitionID), category)
	if err != nil {
		if errors.Is(err, repository.ErrCompetitionNotFound) {
			RespondError(c, http.StatusNotFound, errors.New("competition not found"))
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Build response
	response := models.ParticipantListResponse{
		Participants: make([]*models.ParticipantResponse, len(participants)),
	}

	for i, participant := range participants {
		response.Participants[i] = &models.ParticipantResponse{
			CompetitionID: participant.GetCompetitionID(),
			DossardNumber: participant.GetDossardNumber(),
			FirstName:     participant.GetFirstName(),
			LastName:      participant.GetLastName(),
			Category:      participant.GetCategory(),
			Gender:        participant.GetGender(),
		}
	}

	c.JSON(http.StatusOK, response)
}
