package service

import (
	"context"
	"errors"
	"faceit_stats_wrapper/internal/entity"
	"fmt"
	"time"
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

	const historyTTL = 5 * time.Minute

	needSync := player == nil ||
		player.MatchesSyncedAt == nil ||
		time.Since(*player.MatchesSyncedAt) >= historyTTL

	if needSync {
		if err := s.syncPlayerMatches(ctx, nickname); err != nil {
			return nil, fmt.Errorf("failed to sync player matches: %w", err)
		}

		player, err = s.dbRepoGet.GetPlayerByNickname(ctx, nickname)
		if err != nil {
			return nil, fmt.Errorf("failed to sync player: %w", err)
		}

		if player == nil {
			return nil, fmt.Errorf("player missing after successful sync")
		}
	}

	matches, err := s.dbRepoGet.GetPlayerMatches(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player's matches: %w", err)
	}

	return matches, nil
}

func (s *statsService) syncPlayerMatches(ctx context.Context, nickname string) error {
	player, err := s.GetPlayer(ctx, nickname)
	if err != nil {
		return fmt.Errorf("failed to get player by nickname: %w", err)
	}

	matches, err := s.faceitRepo.GetPlayerMatches(ctx, player.ID, 20)
	if err != nil {
		return fmt.Errorf("failed to get player matches: %w", err)
	}

	for _, v := range matches {
		err = s.dbRepoSet.SaveMatch(ctx, &v, player.ID)
		if err != nil {
			return fmt.Errorf("failed to save match %s: %w", v.MatchID, err)
		}
	}

	if err := s.dbRepoSet.MarkPlayerMatchesSynced(ctx, player.ID); err != nil {
		return fmt.Errorf("failed to update matches sync time: %w", err)
	}

	return nil
}
