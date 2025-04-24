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
