package aggregate

import "github.com/NiskuT/cross-api/internal/domain/entity"

// Run is the aggregate root for run domain
type Run struct {
	run *entity.Run
}

// NewRun creates a new run aggregate
func NewRun() *Run {
	return &Run{run: &entity.Run{}}
}

// GetCompetitionID returns the competition ID
func (r *Run) GetCompetitionID() int32 {
	return r.run.CompetitionID
}

// GetDossard returns the dossard number
func (r *Run) GetDossard() int32 {
	return r.run.Dossard
}

// GetRunNumber returns the run number
func (r *Run) GetRunNumber() int32 {
	return r.run.RunNumber
}

// GetZone returns the zone
func (r *Run) GetZone() string {
	return r.run.Zone
}

// GetDoor1 returns the door1 status
func (r *Run) GetDoor1() bool {
	return r.run.Door1
}

// GetDoor2 returns the door2 status
func (r *Run) GetDoor2() bool {
	return r.run.Door2
}

// GetDoor3 returns the door3 status
func (r *Run) GetDoor3() bool {
	return r.run.Door3
}

// GetDoor4 returns the door4 status
func (r *Run) GetDoor4() bool {
	return r.run.Door4
}

// GetDoor5 returns the door5 status
func (r *Run) GetDoor5() bool {
	return r.run.Door5
}

// GetDoor6 returns the door6 status
func (r *Run) GetDoor6() bool {
	return r.run.Door6
}

// GetPenalty returns the penalty
func (r *Run) GetPenalty() int32 {
	return r.run.Penalty
}

// GetChronoSec returns the chrono seconds
func (r *Run) GetChronoSec() int32 {
	return r.run.ChronoSec
}

// SetCompetitionID sets the competition ID
func (r *Run) SetCompetitionID(competitionID int32) {
	r.run.CompetitionID = competitionID
}

// SetDossard sets the dossard number
func (r *Run) SetDossard(dossard int32) {
	r.run.Dossard = dossard
}

// SetRunNumber sets the run number
func (r *Run) SetRunNumber(runNumber int32) {
	r.run.RunNumber = runNumber
}

// SetZone sets the zone
func (r *Run) SetZone(zone string) {
	r.run.Zone = zone
}

// SetDoor1 sets the door1 status
func (r *Run) SetDoor1(door1 bool) {
	r.run.Door1 = door1
}

// SetDoor2 sets the door2 status
func (r *Run) SetDoor2(door2 bool) {
	r.run.Door2 = door2
}

// SetDoor3 sets the door3 status
func (r *Run) SetDoor3(door3 bool) {
	r.run.Door3 = door3
}

// SetDoor4 sets the door4 status
func (r *Run) SetDoor4(door4 bool) {
	r.run.Door4 = door4
}

// SetDoor5 sets the door5 status
func (r *Run) SetDoor5(door5 bool) {
	r.run.Door5 = door5
}

// SetDoor6 sets the door6 status
func (r *Run) SetDoor6(door6 bool) {
	r.run.Door6 = door6
}

// SetPenalty sets the penalty
func (r *Run) SetPenalty(penalty int32) {
	r.run.Penalty = penalty
}

// SetChronoSec sets the chrono seconds
func (r *Run) SetChronoSec(chronoSec int32) {
	r.run.ChronoSec = chronoSec
}
