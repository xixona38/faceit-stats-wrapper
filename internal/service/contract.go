package service

import (
	"context"
	"faceit_stats_wrapper/internal/entity"
)

type FaceitRepository interface {
	GetPlayerByNickname(ctx context.Context, nickname string) (*entity.Player, error)
	GetPlayerLastMatch(ctx context.Context, playerID string) (*entity.Match, error)
}

type StatsService interface {
	GetPlayer(ctx context.Context, nickname string) (*entity.Player, error)
	GetLastMatch(ctx context.Context, nickname string) (*entity.Match, error)
}
