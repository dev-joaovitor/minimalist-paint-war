// Package persist runs database writes on a background goroutine so the game
// loop never blocks on I/O.
package persist

import (
	"context"
	"log"
	"time"

	"paintwar/server/internal/model"
)

// LeaderboardLimit is how many rows the lobby leaderboard shows.
const LeaderboardLimit = 10

const (
	jobBuffer = 64
	opTimeout = 5 * time.Second
)

// Store is the subset of db.Store the writer needs.
type Store interface {
	UpsertPlayer(ctx context.Context, username string) error
	SaveMatch(ctx context.Context, r model.MatchResult) error
	Leaderboard(ctx context.Context, limit int) ([]model.LeaderEntry, error)
}

type jobKind int

const (
	jobUpsert jobKind = iota
	jobMatch
	jobRefresh
)

type job struct {
	kind     jobKind
	username string
	result   model.MatchResult
}

// Writer drains a job queue, performing DB writes and refreshing the cached
// leaderboard after each change.
type Writer struct {
	store         Store
	jobs          chan job
	onLeaderboard func([]model.LeaderEntry)
}

// New creates a writer over the given store.
func New(store Store) *Writer {
	return &Writer{store: store, jobs: make(chan job, jobBuffer)}
}

// Start launches the worker goroutine and performs an initial leaderboard load.
// onLeaderboard is invoked (on the worker goroutine) whenever the leaderboard
// changes.
func (w *Writer) Start(ctx context.Context, onLeaderboard func([]model.LeaderEntry)) {
	w.onLeaderboard = onLeaderboard
	go w.loop(ctx)
	w.enqueue(job{kind: jobRefresh})
}

// UpsertPlayer queues a player upsert (non-blocking).
func (w *Writer) UpsertPlayer(username string) {
	w.enqueue(job{kind: jobUpsert, username: username})
}

// SaveMatch queues a match write (non-blocking).
func (w *Writer) SaveMatch(r model.MatchResult) {
	w.enqueue(job{kind: jobMatch, result: r})
}

func (w *Writer) enqueue(j job) {
	select {
	case w.jobs <- j:
	default:
		log.Printf("persist: job queue full, dropping %v", j.kind)
	}
}

func (w *Writer) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-w.jobs:
			w.process(ctx, j)
		}
	}
}

func (w *Writer) process(ctx context.Context, j job) {
	opCtx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	switch j.kind {
	case jobUpsert:
		if err := w.store.UpsertPlayer(opCtx, j.username); err != nil {
			log.Printf("persist: upsert %s: %v", j.username, err)
			return
		}
	case jobMatch:
		if err := w.store.SaveMatch(opCtx, j.result); err != nil {
			log.Printf("persist: save match: %v", err)
			return
		}
	case jobRefresh:
		// fall through to refresh only
	}
	w.refresh(opCtx)
}

func (w *Writer) refresh(ctx context.Context) {
	if w.onLeaderboard == nil {
		return
	}
	entries, err := w.store.Leaderboard(ctx, LeaderboardLimit)
	if err != nil {
		log.Printf("persist: leaderboard: %v", err)
		return
	}
	w.onLeaderboard(entries)
}
