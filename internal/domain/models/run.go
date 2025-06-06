package models

// ParticipantInput represents the input for creating a participant
type ParticipantInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	DossardNumber int32  `json:"dossard_number" binding:"required"`
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Category      string `json:"category" binding:"required"`
}

// ParticipantResponse represents the response for a participant
type ParticipantResponse struct {
	CompetitionID int32  `json:"competition_id"`
	DossardNumber int32  `json:"dossard_number"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Category      string `json:"category"`
}

// ParticipantListResponse represents the response for a list of participants
type ParticipantListResponse struct {
	Participants []*ParticipantResponse `json:"participants"`
}

// RunInput represents the input for creating a new run
type RunInput struct {
	CompetitionID int32  `json:"competition_id" binding:"required"`
	Dossard       int32  `json:"dossard" binding:"required"`
	Zone          string `json:"zone" binding:"required"`
	Door1         bool   `json:"door1"`
	Door2         bool   `json:"door2"`
	Door3         bool   `json:"door3"`
	Door4         bool   `json:"door4"`
	Door5         bool   `json:"door5"`
	Door6         bool   `json:"door6"`
	Penality      int32  `json:"penality"`
	ChronoSec     int32  `json:"chrono_sec"`
}

// RunResponse represents the response for a run
type RunResponse struct {
	CompetitionID int32  `json:"competition_id"`
	Dossard       int32  `json:"dossard"`
	RunNumber     int32  `json:"run_number"`
	Zone          string `json:"zone"`
	Door1         bool   `json:"door1"`
	Door2         bool   `json:"door2"`
	Door3         bool   `json:"door3"`
	Door4         bool   `json:"door4"`
	Door5         bool   `json:"door5"`
	Door6         bool   `json:"door6"`
	Penality      int32  `json:"penality"`
	ChronoSec     int32  `json:"chrono_sec"`
}
