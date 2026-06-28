package persist

import (
	"context"
	"sync"
	"testing"
	"time"

	"paintwar/server/internal/model"
)

type fakeStore struct {
	mu      sync.Mutex
	upserts []string
	matches int
}

func (f *fakeStore) UpsertPlayer(_ context.Context, u string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upserts = append(f.upserts, u)
	return nil
}

func (f *fakeStore) SaveMatch(_ context.Context, _ model.MatchResult) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.matches++
	return nil
}

func (f *fakeStore) Leaderboard(_ context.Context, _ int) ([]model.LeaderEntry, error) {
	return []model.LeaderEntry{{Username: "alice", Wins: 3, Losses: 1}}, nil
}

func TestWriterRefreshesLeaderboard(t *testing.T) {
	fake := &fakeStore{}
	w := New(fake)

	got := make(chan []model.LeaderEntry, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx, func(e []model.LeaderEntry) { got <- e })

	// Initial refresh fires on Start.
	select {
	case e := <-got:
		if len(e) != 1 || e[0].Username != "alice" {
			t.Fatalf("unexpected leaderboard %+v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no initial leaderboard refresh")
	}

	// A match write triggers another refresh.
	w.SaveMatch(model.MatchResult{Winner: model.TeamRed})
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("no refresh after SaveMatch")
	}

	fake.mu.Lock()
	matches := fake.matches
	fake.mu.Unlock()
	if matches != 1 {
		t.Fatalf("expected 1 match saved, got %d", matches)
	}
}
