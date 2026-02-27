package fpl

import "fmt"

const (
	posGK  = 1
	posDEF = 2
	posMID = 3
	posFWD = 4
)

// ValidateStartingTeam checks whether the given player IDs form a legal FPL starting XI.
func (s *Store) ValidateStartingTeam(playerIDs []int) error {
	if len(playerIDs) != 11 {
		return fmt.Errorf("starting team must have 11 players, got %d", len(playerIDs))
	}

	seen := make(map[int]struct{}, 11)
	for _, id := range playerIDs {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("duplicate player id %d", id)
		}
		seen[id] = struct{}{}
	}

	counts := make(map[int]int) // position -> count
	clubCounts := make(map[int]int) // team id -> count

	for _, id := range playerIDs {
		p, ok := s.Players[id]
		if !ok {
			return fmt.Errorf("unknown player id %d", id)
		}
		counts[p.ElementType]++
		clubCounts[p.Team]++
	}

	for teamID, n := range clubCounts {
		if n > 3 {
			return fmt.Errorf("too many players from team %s (%d, max 3)", s.TeamName(teamID), n)
		}
	}

	gk := counts[posGK]
	def := counts[posDEF]
	mid := counts[posMID]
	fwd := counts[posFWD]

	if gk != 1 {
		return fmt.Errorf("must have exactly 1 GK, got %d", gk)
	}
	if def < 3 || def > 5 {
		return fmt.Errorf("must have 3-5 DEF, got %d", def)
	}
	if mid < 2 || mid > 5 {
		return fmt.Errorf("must have 2-5 MID, got %d", mid)
	}
	if fwd < 1 || fwd > 3 {
		return fmt.Errorf("must have 1-3 FWD, got %d", fwd)
	}

	return nil
}

// TeamEventPoints returns the sum of current gameweek points for the given player IDs.
// Call ValidateStartingTeam first to ensure the IDs are valid.
func (s *Store) TeamEventPoints(playerIDs []int) int {
	var total int
	for _, id := range playerIDs {
		total += s.Players[id].EventPoints
	}
	return total
}
