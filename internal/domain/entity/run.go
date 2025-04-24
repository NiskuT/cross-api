package entity

type Run struct {
	CompetitionID int32
	Dossard       int32
	RunNumber     int32
	Zone          string
	Door1         bool
	Door2         bool
	Door3         bool
	Door4         bool
	Door5         bool
	Door6         bool
	Penalty       int32
	ChronoSec     int32
}
