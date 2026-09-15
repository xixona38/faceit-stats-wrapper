package service

import (
	"context"
	"faceit_stats_wrapper/internal/entity"
)

type FaceitRepository interface {
	GetPlayerByNickname(ctx context.Context, nickname string) (*entity.Player, error)
	GetPlayerLastMatch(ctx context.Context, playerID string) (*entity.Match, error)
	GetPlayerMatches(ctx context.Context, playerID string, limit int) ([]entity.Match, error)
}

type StatsService interface {
	GetPlayer(ctx context.Context, nickname string) (*entity.Player, error)
	GetLastMatch(ctx context.Context, nickname string) (*entity.Match, error)
	GetPlayerMatches(ctx context.Context, nickname string) ([]entity.Match, error)
	SyncPlayerMatches(ctx context.Context, nickname string) error
}

type DBRepositorySetter interface {
	SavePlayer(ctx context.Context, player *entity.Player) error
	SaveMatch(ctx context.Context, match *entity.Match, playerID string) error
}

type DBRepositoryGetter interface {
	GetPlayerByNickname(ctx context.Context, nickname string) (*entity.Player, error)
	GetPlayerMatches(ctx context.Context, playerID string) ([]entity.Match, error)
}
