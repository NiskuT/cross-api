package models

// ParticipantResponse represents the response for a participant
type ParticipantResponse struct {
	CompetitionID int32  `json:"competition_id"`
	DossardNumber int32  `json:"dossard_number"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Category      string `json:"category"`
}
