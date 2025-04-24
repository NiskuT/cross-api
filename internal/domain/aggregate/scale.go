package aggregate

import "github.com/NiskuT/cross-api/internal/domain/entity"

type Scale struct {
	scale *entity.Scale
}

func NewScale() *Scale {
	return &Scale{
		scale: &entity.Scale{},
	}
}

func (s *Scale) GetCompetitionID() int32 {
	return s.scale.CompetitionID
}

func (s *Scale) GetCategory() string {
	return s.scale.Category
}

func (s *Scale) GetPointsDoor1() int32 {
	return s.scale.PointsDoor1
}

func (s *Scale) GetPointsDoor2() int32 {
	return s.scale.PointsDoor2
}

func (s *Scale) GetPointsDoor3() int32 {
	return s.scale.PointsDoor3
}

func (s *Scale) GetPointsDoor4() int32 {
	return s.scale.PointsDoor4
}

func (s *Scale) GetPointsDoor5() int32 {
	return s.scale.PointsDoor5
}

func (s *Scale) GetPointsDoor6() int32 {
	return s.scale.PointsDoor6
}

func (s *Scale) SetCompetitionID(competitionID int32) {
	s.scale.CompetitionID = competitionID
}

func (s *Scale) SetCategory(category string) {
	s.scale.Category = category
}

func (s *Scale) SetPointsDoor1(points int32) {
	s.scale.PointsDoor1 = points
}

func (s *Scale) SetPointsDoor2(points int32) {
	s.scale.PointsDoor2 = points
}

func (s *Scale) SetPointsDoor3(points int32) {
	s.scale.PointsDoor3 = points
}

func (s *Scale) SetPointsDoor4(points int32) {
	s.scale.PointsDoor4 = points
}

func (s *Scale) SetPointsDoor5(points int32) {
	s.scale.PointsDoor5 = points
}

func (s *Scale) SetPointsDoor6(points int32) {
	s.scale.PointsDoor6 = points
}
