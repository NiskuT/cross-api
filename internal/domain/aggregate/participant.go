package aggregate

import "github.com/NiskuT/cross-api/internal/domain/entity"

type Participant struct {
	participant *entity.Participant
}

func NewParticipant() *Participant {
	return &Participant{participant: &entity.Participant{}}
}

func (p *Participant) GetCompetitionID() int32 {
	return p.participant.CompetitionID
}

func (p *Participant) GetDossardNumber() int32 {
	return p.participant.DossardNumber
}

func (p *Participant) GetFirstName() string {
	return p.participant.FirstName
}

func (p *Participant) GetLastName() string {
	return p.participant.LastName
}

func (p *Participant) GetCategory() string {
	return p.participant.Category
}

func (p *Participant) SetCompetitionID(competitionID int32) {
	p.participant.CompetitionID = competitionID
}

func (p *Participant) SetDossardNumber(dossardNumber int32) {
	p.participant.DossardNumber = dossardNumber
}

func (p *Participant) SetFirstName(firstName string) {
	p.participant.FirstName = firstName
}

func (p *Participant) SetLastName(lastName string) {
	p.participant.LastName = lastName
}

func (p *Participant) SetCategory(category string) {
	p.participant.Category = category
}
