package entity

// Competition represents a competition entity
type Competition struct {
	ID          int32
	Name        string
	Description string
	Date        string
	Location    string
	Organizer   string
	Contact     string
}
