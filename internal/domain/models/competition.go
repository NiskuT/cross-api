package models

type Competition struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Date        string `json:"date,omitempty"`
	Location    string `json:"location,omitempty"`
	Organizer   string `json:"organizer,omitempty"`
	Contact     string `json:"contact,omitempty"`
}

type CompetitionResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Organizer   string `json:"organizer"`
	Contact     string `json:"contact"`
}

type CompetitionListResponse struct {
	Competitions []*CompetitionResponse `json:"competitions"`
}

type CompetitionScaleInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	Category      string `json:"category" binding:"required"`
	Zone          string `json:"zone" binding:"required"`
	PointsDoor1   int32  `json:"points_door1" binding:"required"`
	PointsDoor2   int32  `json:"points_door2" binding:"required"`
	PointsDoor3   int32  `json:"points_door3" binding:"required"`
	PointsDoor4   int32  `json:"points_door4" binding:"required"`
	PointsDoor5   int32  `json:"points_door5" binding:"required"`
	PointsDoor6   int32  `json:"points_door6" binding:"required"`
}

// CompetitionZoneDeleteInput represents the input for deleting a zone from a competition
type CompetitionZoneDeleteInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	Category      string `json:"category" binding:"required"`
	Zone          string `json:"zone" binding:"required"`
}

// RefereeInput represents the input for adding a referee to a competition
type RefereeInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
}

// RefereeInvitationResponse represents the response for generating a referee invitation link
type RefereeInvitationResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// RefereeInvitationAcceptInput represents the input for accepting a referee invitation
type RefereeInvitationAcceptInput struct {
	Token string `json:"token" binding:"required"`
}

// RefereeInvitationAcceptUnauthenticatedInput represents the input for accepting a referee invitation without authentication
type RefereeInvitationAcceptUnauthenticatedInput struct {
	Token     string `json:"token" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
}

// ZoneResponse represents a single zone in a competition
type ZoneResponse struct {
	Zone        string `json:"zone"`
	Category    string `json:"category"`
	PointsDoor1 int32  `json:"points_door1"`
	PointsDoor2 int32  `json:"points_door2"`
	PointsDoor3 int32  `json:"points_door3"`
	PointsDoor4 int32  `json:"points_door4"`
	PointsDoor5 int32  `json:"points_door5"`
	PointsDoor6 int32  `json:"points_door6"`
}

// ZonesListResponse represents a list of zones in a competition
type ZonesListResponse struct {
	CompetitionID int32          `json:"competition_id"`
	Zones         []ZoneResponse `json:"zones"`
}

// LiverankingResponse represents a single liveranking entry
type LiverankingResponse struct {
	Rank         int32  `json:"rank"`
	Dossard      int32  `json:"dossard"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Category     string `json:"category"`
	Gender       string `json:"gender"`
	NumberOfRuns int32  `json:"number_of_runs"`
	TotalPoints  int32  `json:"total_points"`
	Penality     int32  `json:"penality"`
	ChronoSec    int32  `json:"chrono_sec"`
}

// LiverankingListResponse represents a list of liveranking entries
type LiverankingListResponse struct {
	CompetitionID int32                 `json:"competition_id"`
	Category      string                `json:"category,omitempty"`
	Gender        string                `json:"gender,omitempty"`
	Page          int32                 `json:"page"`
	PageSize      int32                 `json:"page_size"`
	Total         int32                 `json:"total"`
	Rankings      []LiverankingResponse `json:"rankings"`
}
