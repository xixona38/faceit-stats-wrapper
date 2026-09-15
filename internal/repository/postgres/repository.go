package postgres

import (
	"context"
	"errors"
	"faceit_stats_wrapper/internal/entity"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepoSet struct {
	pool *pgxpool.Pool
}

func NewPostgresRepoSet(pool *pgxpool.Pool) *PostgresRepoSet {
	return &PostgresRepoSet{
		pool: pool,
	}
}

type PostgresRepoGet struct {
	pool *pgxpool.Pool
}

func NewPostgresRepoGet(pool *pgxpool.Pool) *PostgresRepoGet {
	return &PostgresRepoGet{
		pool: pool,
	}
}

func (r *PostgresRepoSet) SavePlayer(ctx context.Context, player *entity.Player) error {
	query := `
		INSERT INTO players (player_id, nickname)
		VALUES ($1, $2)
		ON CONFLICT (player_id) DO UPDATE
		SET nickname = EXCLUDED.nickname, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.pool.Exec(ctx, query, player.ID, player.Nickname)
	if err != nil {
		return fmt.Errorf("failed to save player: %w", err)
	}

	return nil
}

func (r *PostgresRepoSet) SaveMatch(ctx context.Context, match *entity.Match, playerID string) error {
	query := `
		INSERT INTO matches (match_id, player_id, map, result, score, kills, deaths, kd_ratio)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (match_id, player_id) DO NOTHING
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		match.MatchID,
		playerID,
		match.Map,
		match.Result,
		match.Score,
		match.Kills,
		match.Deaths,
		match.KDRatio,
	)
	if err != nil {
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}

func (r *PostgresRepoGet) GetPlayerByNickname(ctx context.Context, nickname string) (*entity.Player, error) {
	query := `
		SELECT player_id, nickname, matches_synced_at 
		FROM players 
		WHERE nickname = $1
	`
	var player entity.Player
	err := r.pool.QueryRow(ctx, query, nickname).Scan(&player.ID, &player.Nickname, &player.MatchesSyncedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get player by nickname: %w", err)
	}

	return &player, nil
}

func (r *PostgresRepoGet) GetPlayerMatches(ctx context.Context, playerID string) ([]entity.Match, error) {
	query := `
		SELECT match_id, map, result, score, kills, deaths, kd_ratio
		FROM matches
		WHERE player_id = $1
		ORDER BY saved_at DESC
		LIMIT 20
	`

	rows, err := r.pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to postgres: %w", err)
	}
	defer rows.Close()

	matches := make([]entity.Match, 0, 10)

	for rows.Next() {
		var match entity.Match

		err := rows.Scan(
			&match.MatchID,
			&match.Map,
			&match.Result,
			&match.Score,
			&match.Kills,
			&match.Deaths,
			&match.KDRatio,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get match: %w", err)
		}

		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occured while executing the a query: %w", err)
	}

	return matches, nil
}
