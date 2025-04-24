package aggregate

import "github.com/NiskuT/cross-api/internal/domain/entity"

// Competition is the aggregate root for competition domain
type Competition struct {
	competition *entity.Competition
}

// NewCompetition creates a new competition aggregate
func NewCompetition() *Competition {
	return &Competition{
		competition: &entity.Competition{},
	}
}

// GetID returns the competition ID
func (c *Competition) GetID() int32 {
	return c.competition.ID
}

// GetName returns the competition name
func (c *Competition) GetName() string {
	return c.competition.Name
}

// GetDescription returns the competition description
func (c *Competition) GetDescription() string {
	return c.competition.Description
}

// GetDate returns the competition date
func (c *Competition) GetDate() string {
	return c.competition.Date
}

// GetLocation returns the competition location
func (c *Competition) GetLocation() string {
	return c.competition.Location
}

// GetOrganizer returns the competition organizer
func (c *Competition) GetOrganizer() string {
	return c.competition.Organizer
}

// GetContact returns the competition contact
func (c *Competition) GetContact() string {
	return c.competition.Contact
}

// SetID sets the competition ID
func (c *Competition) SetID(id int32) {
	c.competition.ID = id
}

// SetName sets the competition name
func (c *Competition) SetName(name string) {
	c.competition.Name = name
}

// SetDescription sets the competition description
func (c *Competition) SetDescription(description string) {
	c.competition.Description = description
}

// SetDate sets the competition date
func (c *Competition) SetDate(date string) {
	c.competition.Date = date
}

// SetLocation sets the competition location
func (c *Competition) SetLocation(location string) {
	c.competition.Location = location
}

// SetOrganizer sets the competition organizer
func (c *Competition) SetOrganizer(organizer string) {
	c.competition.Organizer = organizer
}

// SetContact sets the competition contact
func (c *Competition) SetContact(contact string) {
	c.competition.Contact = contact
}
