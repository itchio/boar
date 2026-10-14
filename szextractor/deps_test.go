package szextractor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/itchio/boar/szextractor/types"
	"github.com/itchio/headway/state"
)

func TestHealFormulaDepsReturnsCancellation(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	spec := &types.DepSpec{Sources: []string{server.URL + "/a.zip", server.URL + "/b.zip"}}
	err := healFormulaDeps(ctx, &state.Consumer{}, spec, []types.DepEntry{{Name: "x"}}, t.TempDir())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("expected no source to be tried after cancel, got %d requests", n)
	}
}

func TestHealFormulaDepsFailsWhenAllSourcesFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	spec := &types.DepSpec{Sources: []string{server.URL + "/a.zip", server.URL + "/b.zip"}}
	err := healFormulaDeps(context.Background(), &state.Consumer{}, spec, []types.DepEntry{{Name: "x"}}, t.TempDir())
	if err == nil {
		t.Fatal("expected an error when every source fails")
	}
}
