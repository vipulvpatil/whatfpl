package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	bootstrapURL   = "https://fantasy.premierleague.com/api/bootstrap-static/"
	interval       = 100 * time.Millisecond
	maxConcurrent  = 20
	requestTimeout = 3 * time.Second
)

func main() {
	targetsFlag := flag.String("targets", "http://localhost:8080", "comma-separated list of server base URLs")
	flag.Parse()

	targets := strings.Split(*targetsFlag, ",")
	for i, t := range targets {
		targets[i] = strings.TrimRight(t, "/") + "/players"
	}

	fplClient := &http.Client{Timeout: 15 * time.Second}

	players, err := fetchPlayers(fplClient)
	if err != nil {
		log.Fatalf("failed to fetch FPL data: %v", err)
	}

	entries := buildEntries(players)
	log.Printf("checker started: %d entries, firing every %s", len(entries), interval)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	client := &http.Client{Timeout: requestTimeout}
	sem := make(chan struct{}, maxConcurrent)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		n := rng.Intn(10) + 1 // 1–10 concurrent requests per tick
		for range n {
			select {
			case sem <- struct{}{}:
				ids := entries[rng.Intn(len(entries))]
				target := targets[rng.Intn(len(targets))]
				go func() {
					defer func() { <-sem }()
					call(client, target, ids)
				}()
			default:
				// at capacity, skip
			}
		}
	}
}

func call(client *http.Client, target string, ids []int) {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}

	resp, err := client.Get(target + "?ids=" + strings.Join(parts, ","))
	if err != nil {
		log.Printf("ERR %v", err)
		return
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		log.Printf("ERR status=%d", resp.StatusCode)
	}
}

// player holds only what the checker needs from the FPL bootstrap.
type player struct {
	ID          int `json:"id"`
	Team        int `json:"team"`
	ElementType int `json:"element_type"` // 1=GK 2=DEF 3=MID 4=FWD
}

func fetchPlayers(client *http.Client) ([]player, error) {
	resp, err := client.Get(bootstrapURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Elements []player `json:"elements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Elements, nil
}

// buildEntries generates 500+ entries using real FPL player data.
func buildEntries(players []player) [][]int {
	rng := rand.New(rand.NewSource(42))

	byPos := map[int][]int{1: {}, 2: {}, 3: {}, 4: {}}
	byTeam := map[int][]int{}
	for _, p := range players {
		byPos[p.ElementType] = append(byPos[p.ElementType], p.ID)
		byTeam[p.Team] = append(byTeam[p.Team], p.ID)
	}

	out := make([][]int, 0, 1000)

	// 960 valid: real IDs in a legal formation (1 GK, 4 DEF, 4 MID, 2 FWD)
	for range 960 {
		out = append(out, validTeam(rng, byPos))
	}

	// 5 invalid: too few players (1–10)
	for range 5 {
		n := rng.Intn(10) + 1
		out = append(out, sample(rng, players, n))
	}

	// 5 invalid: too many players (12–16)
	for range 5 {
		n := rng.Intn(5) + 12
		out = append(out, sample(rng, players, n))
	}

	// 5 invalid: duplicate IDs (10 distinct + first repeated)
	for range 5 {
		ids := sample(rng, players, 10)
		out = append(out, append(ids, ids[0]))
	}

	// 5 invalid: non-existent player IDs
	for range 5 {
		ids := make([]int, 11)
		for i := range ids {
			ids[i] = 90000 + rng.Intn(9999)
		}
		out = append(out, ids)
	}

	// 4 invalid: >3 players from the same team
	for range 4 {
		out = append(out, tooManyFromOneTeam(rng, byPos, byTeam))
	}

	// 4 invalid: 0 GKs
	for range 4 {
		ids := pick(rng, byPos[2], 4)
		ids = append(ids, pick(rng, byPos[3], 5)...)
		ids = append(ids, pick(rng, byPos[4], 2)...)
		out = append(out, ids)
	}

	// 4 invalid: 2 GKs
	for range 4 {
		ids := pick(rng, byPos[1], 2)
		ids = append(ids, pick(rng, byPos[2], 4)...)
		ids = append(ids, pick(rng, byPos[3], 3)...)
		ids = append(ids, pick(rng, byPos[4], 2)...)
		out = append(out, ids)
	}

	// 4 invalid: only 2 DEF
	for range 4 {
		ids := pick(rng, byPos[1], 1)
		ids = append(ids, pick(rng, byPos[2], 2)...)
		ids = append(ids, pick(rng, byPos[3], 5)...)
		ids = append(ids, pick(rng, byPos[4], 3)...)
		out = append(out, ids)
	}

	// 4 invalid: 0 FWD
	for range 4 {
		ids := pick(rng, byPos[1], 1)
		ids = append(ids, pick(rng, byPos[2], 5)...)
		ids = append(ids, pick(rng, byPos[3], 5)...)
		out = append(out, ids)
	}

	return out
}

// validTeam builds a legal 1-4-4-2 team from real player IDs.
func validTeam(rng *rand.Rand, byPos map[int][]int) []int {
	ids := pick(rng, byPos[1], 1)
	ids = append(ids, pick(rng, byPos[2], 4)...)
	ids = append(ids, pick(rng, byPos[3], 4)...)
	ids = append(ids, pick(rng, byPos[4], 2)...)
	return ids
}

// tooManyFromOneTeam builds an 11-player team with 4 players from one club.
func tooManyFromOneTeam(rng *rand.Rand, byPos map[int][]int, byTeam map[int][]int) []int {
	// Find a team with enough players
	var clubIDs []int
	for _, ids := range byTeam {
		if len(ids) >= 4 {
			clubIDs = ids
			break
		}
	}
	ids := pick(rng, clubIDs, 4)
	// Fill remaining 7 spots avoiding position constraints — just use any valid formation
	remaining := validTeam(rng, byPos)
	// Deduplicate: remove from remaining any ID already in ids
	inClub := make(map[int]bool, 4)
	for _, id := range ids {
		inClub[id] = true
	}
	for _, id := range remaining {
		if !inClub[id] {
			ids = append(ids, id)
		}
		if len(ids) == 11 {
			break
		}
	}
	return ids
}

// pick returns n randomly sampled (with replacement) IDs from src.
func pick(rng *rand.Rand, src []int, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = src[rng.Intn(len(src))]
	}
	return out
}

// sample returns n IDs drawn from the full player list (with replacement).
func sample(rng *rand.Rand, players []player, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = players[rng.Intn(len(players))].ID
	}
	return out
}
