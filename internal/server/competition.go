package server

import (
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
