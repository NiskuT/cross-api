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

// RefereeInput represents the input for adding a referee to a competition
type RefereeInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
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
