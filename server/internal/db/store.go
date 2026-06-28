package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"paintwar/server/internal/model"
)

//go:embed migrations/001_init.sql
var initSchema string

// Store reads and writes player stats and match history.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore wraps a connection pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Migrate applies the (idempotent) schema.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, initSchema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// UpsertPlayer ensures a player row exists and refreshes last_seen.
func (s *Store) UpsertPlayer(ctx context.Context, username string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO players (username) VALUES ($1)
		 ON CONFLICT (username) DO UPDATE SET last_seen = now()`,
		username)
	return err
}

// Leaderboard returns the top players ranked by wins desc, losses asc.
func (s *Store) Leaderboard(ctx context.Context, limit int) ([]model.LeaderEntry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT username, wins, losses FROM players
		 ORDER BY wins DESC, losses ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LeaderEntry
	for rows.Next() {
		var e model.LeaderEntry
		if err := rows.Scan(&e.Username, &e.Wins, &e.Losses); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SaveMatch records a finished match and updates win/loss tallies in one
// transaction. Draws do not change win/loss counts.
func (s *Store) SaveMatch(ctx context.Context, r model.MatchResult) error {
	winner := string(r.Winner)
	if r.Winner == model.TeamNone {
		winner = "DRAW"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	var matchID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO matches (seed, red_score, green_score, winner, duration_ms)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		r.Seed, r.Red, r.Green, winner, r.DurationMs,
	).Scan(&matchID); err != nil {
		return err
	}

	for _, p := range r.Players {
		// Guard against a player who never had a row (shouldn't happen).
		if _, err := tx.Exec(ctx,
			`INSERT INTO players (username) VALUES ($1) ON CONFLICT DO NOTHING`,
			p.Username); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO match_players (match_id, username, team, result)
			 VALUES ($1, $2, $3, $4)`,
			matchID, p.Username, string(p.Team), p.Result); err != nil {
			return err
		}
		switch p.Result {
		case "win":
			if _, err := tx.Exec(ctx, `UPDATE players SET wins = wins + 1 WHERE username = $1`, p.Username); err != nil {
				return err
			}
		case "loss":
			if _, err := tx.Exec(ctx, `UPDATE players SET losses = losses + 1 WHERE username = $1`, p.Username); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
