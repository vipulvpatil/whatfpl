package fpl

import "fmt"

type Store struct {
	Players        map[int]Player
	Teams          map[int]Team
	ElementTypes   map[int]ElementType
	CurrentGameweek int
}

func NewStore(data *bootstrapResponse) (*Store, error) {
	s := &Store{
		Players:      make(map[int]Player, len(data.Elements)),
		Teams:        make(map[int]Team, len(data.Teams)),
		ElementTypes: make(map[int]ElementType, len(data.ElementTypes)),
	}

	for _, p := range data.Elements {
		s.Players[p.ID] = p
	}
	for _, t := range data.Teams {
		s.Teams[t.ID] = t
	}
	for _, et := range data.ElementTypes {
		s.ElementTypes[et.ID] = et
	}

	for _, e := range data.Events {
		if e.IsCurrent {
			s.CurrentGameweek = e.ID
			break
		}
	}
	if s.CurrentGameweek == 0 {
		// Fall back to the last event before "next"
		for _, e := range data.Events {
			if e.IsNext && e.ID > 1 {
				s.CurrentGameweek = e.ID - 1
				break
			}
		}
	}
	if s.CurrentGameweek == 0 {
		return nil, fmt.Errorf("could not determine current gameweek from events")
	}

	return s, nil
}

func (s *Store) PositionName(elementType int) string {
	if et, ok := s.ElementTypes[elementType]; ok {
		return et.SingularNameShort
	}
	return "UNK"
}

func (s *Store) TeamName(teamID int) string {
	if t, ok := s.Teams[teamID]; ok {
		return t.Name
	}
	return "Unknown"
}
