package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/models"
	"github.com/NiskuT/cross-api/internal/server/middlewares"
	"github.com/gin-gonic/gin"
)

// createCompetition godoc
// @Summary      Create a competition
// @Description  Creates a new competition and returns a JWT token.
// @Tags         competition
// @Accept       json
// @Produce      json
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

	c.SetCookie(AccessToken, newToken.GetAccessToken(), 0, "/", "", true, true)
	c.SetCookie(RefreshToken, newToken.GetRefreshToken(), 0, "/", "", true, true)

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

	err := s.competitionService.AddZone(c, competitionScaleInput.CompetitionID, scale)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone added to competition"})
}
