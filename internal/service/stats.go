package service

import (
	"context"
	"errors"
	"faceit_stats_wrapper/internal/entity"
	"fmt"
)

type statsService struct {
	faceitRepo FaceitRepository
	dbRepoSet  DBRepositorySetter
	dbRepoGet  DBRepositoryGetter
}

func NewStatsService(
	faceitRepo FaceitRepository,
	dbRepoSet DBRepositorySetter,
	dbRepoGet DBRepositoryGetter,
) StatsService {
	return &statsService{
		faceitRepo: faceitRepo,
		dbRepoSet:  dbRepoSet,
		dbRepoGet:  dbRepoGet,
	}
}

var (
	ErrPlayerNotFound = errors.New("player not found")
)

func (s *statsService) GetPlayer(ctx context.Context, nickname string) (*entity.Player, error) {
	player, err := s.faceitRepo.GetPlayerByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to get player %s: %w", nickname, err)
	}

	err = s.dbRepoSet.SavePlayer(ctx, player)
	if err != nil {
		return nil, fmt.Errorf("failed to save player in postgres: %w", err)
	}

	return player, nil
}

func (s *statsService) GetLastMatch(ctx context.Context, nickname string) (*entity.Match, error) {
	player, err := s.GetPlayer(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to get player ID for match lookup: %w", err)
	}

	match, err := s.faceitRepo.GetPlayerLastMatch(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load the last match for %s: %w", nickname, err)
	}

	err = s.dbRepoSet.SaveMatch(ctx, match, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to save match in postgres: %w", err)
	}

	return match, nil
}

func (s *statsService) GetPlayerMatches(ctx context.Context, nickname string) ([]entity.Match, error) {
	player, err := s.dbRepoGet.GetPlayerByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if player == nil {
		return nil, ErrPlayerNotFound
	}

	matches, err := s.dbRepoGet.GetPlayerMatches(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player's matches: %w", err)
	}

	return matches, nil
}
