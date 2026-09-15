package entity

import "time"

type Player struct {
	ID              string     `json:"player_id"`
	Nickname        string     `json:"nickname"`
	Level           int        `json:"level"`
	Elo             int        `json:"elo"`
	Avatar          string     `json:"avatar_url,omitempty"`
	MatchesSyncedAt *time.Time `json:"-"`
}
