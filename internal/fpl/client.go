package fpl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const bootstrapURL = "https://fantasy.premierleague.com/api/bootstrap-static/"

type bootstrapResponse struct {
	Elements     []Player      `json:"elements"`
	Teams        []Team        `json:"teams"`
	Events       []Event       `json:"events"`
	ElementTypes []ElementType `json:"element_types"`
}

type Player struct {
	ID          int     `json:"id"`
	WebName     string  `json:"web_name"`
	ElementType int     `json:"element_type"` // 1=GK 2=DEF 3=MID 4=FWD
	Team        int     `json:"team"`
	NowCost     int     `json:"now_cost"` // price × 10
	EventPoints int     `json:"event_points"`
	TotalPoints int     `json:"total_points"`
	Status      string  `json:"status"` // a=available, i=injured, u=unavailable
	Form        string  `json:"form"`
}

type Team struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type Event struct {
	ID        int  `json:"id"`
	IsCurrent bool `json:"is_current"`
	IsNext    bool `json:"is_next"`
}

type ElementType struct {
	ID              int    `json:"id"`
	SingularName    string `json:"singular_name"`
	SingularNameShort string `json:"singular_name_short"`
}

func FetchBootstrap() (*bootstrapResponse, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(bootstrapURL)
	if err != nil {
		return nil, fmt.Errorf("fetching bootstrap-static: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bootstrap-static returned status %d", resp.StatusCode)
	}

	var data bootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding bootstrap-static: %w", err)
	}
	return &data, nil
}
