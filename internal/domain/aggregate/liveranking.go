package aggregate

import "github.com/NiskuT/cross-api/internal/domain/entity"

type Liveranking struct {
	participant  *entity.Participant
	numberOfRuns int32
	totalPoints  int32
	penality     int32
	chronoSec    int32
}

func NewLiveranking() *Liveranking {
	return &Liveranking{
		participant: &entity.Participant{},
	}
}

func (l *Liveranking) GetCompetitionID() int32 {
	return l.participant.CompetitionID
}

func (l *Liveranking) GetDossard() int32 {
	return l.participant.DossardNumber
}

func (l *Liveranking) GetFirstName() string {
	return l.participant.FirstName
}

func (l *Liveranking) GetLastName() string {
	return l.participant.LastName
}

func (l *Liveranking) GetCategory() string {
	return l.participant.Category
}

func (l *Liveranking) GetGender() string {
	return l.participant.Gender
}

func (l *Liveranking) GetNumberOfRuns() int32 {
	return l.numberOfRuns
}

func (l *Liveranking) GetTotalPoints() int32 {
	return l.totalPoints
}

func (l *Liveranking) GetPenality() int32 {
	return l.penality
}

func (l *Liveranking) GetChronoSec() int32 {
	return l.chronoSec
}

func (l *Liveranking) SetCompetitionID(competitionID int32) {
	l.participant.CompetitionID = competitionID
}

func (l *Liveranking) SetDossard(dossard int32) {
	l.participant.DossardNumber = dossard
}

func (l *Liveranking) SetFirstName(firstName string) {
	l.participant.FirstName = firstName
}

func (l *Liveranking) SetLastName(lastName string) {
	l.participant.LastName = lastName
}

func (l *Liveranking) SetCategory(category string) {
	l.participant.Category = category
}

func (l *Liveranking) SetGender(gender string) {
	l.participant.Gender = gender
}

func (l *Liveranking) SetNumberOfRuns(numberOfRuns int32) {
	l.numberOfRuns = numberOfRuns
}

func (l *Liveranking) SetTotalPoints(totalPoints int32) {
	l.totalPoints = totalPoints
}

func (l *Liveranking) SetPenality(penality int32) {
	l.penality = penality
}

func (l *Liveranking) SetChronoSec(chronoSec int32) {
	l.chronoSec = chronoSec
}
