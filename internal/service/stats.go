package service

import (
	"context"
	"faceit_stats_wrapper/internal/entity"
	"fmt"
)

type statsService struct {
	repo FaceitRepository
}

func NewStatsService(repo FaceitRepository) StatsService {
	return &statsService{
		repo: repo,
	}
}

func (s *statsService) GetPlayer(ctx context.Context, nickname string) (*entity.Player, error) {
	player, err := s.repo.GetPlayerByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to get player %s: %w", nickname, err)
	}

	return player, nil
}

func (s *statsService) GetLastMatch(ctx context.Context, nickname string) (*entity.Match, error) {
	player, err := s.repo.GetPlayerByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to get player ID for match lookup: %w", err)
	}

	match, err := s.repo.GetPlayerLastMatch(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load the last match for %s: %w", nickname, err)
	}

	return match, nil
}
