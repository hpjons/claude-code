package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"golog/internal/model"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestLogLifecycle(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	p, err := s.CreateProject(ctx, "Reykjanes")
	if err != nil {
		t.Fatal(err)
	}
	l, err := s.CreateLog(ctx, p.ID, "Section A")
	if err != nil {
		t.Fatal(err)
	}
	if l.Revision != 1 || len(l.Beds) != 0 {
		t.Fatalf("new log = rev %d, %d beds", l.Revision, len(l.Beds))
	}

	phi := 2.5
	l.Beds = append(l.Beds, model.Bed{ID: "b1", Thickness: 1.2, Contact: "sharp",
		Lithology: []model.Lithology{{Pattern: "sandstone", Percent: 100}}, Grain: model.Grain{Base: &phi}})
	if err := s.UpdateLog(ctx, l); err != nil {
		t.Fatal(err)
	}
	if l.Revision != 2 {
		t.Fatalf("revision after save = %d, want 2", l.Revision)
	}

	got, err := s.Log(ctx, l.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Beds) != 1 || got.Beds[0].Thickness != 1.2 || *got.Beds[0].Grain.Base != 2.5 {
		t.Fatalf("read back %+v", got.Beds)
	}

	stale := *got
	stale.Revision = 1
	if err := s.UpdateLog(ctx, &stale); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale save: err = %v, want ErrConflict", err)
	}

	proj, err := s.Project(ctx, p.ID)
	if err != nil || len(proj.Logs) != 1 || proj.Logs[0].BedCount != 1 {
		t.Fatalf("project = %+v, err %v", proj, err)
	}

	if err := s.DeleteProject(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Log(ctx, l.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("log after project delete: err = %v, want ErrNotFound", err)
	}
}

func TestCreateLogNeedsProject(t *testing.T) {
	if _, err := open(t).CreateLog(context.Background(), "nope", "x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestReopenKeepsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.CreateProject(context.Background(), "Kept")
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	list, _ := s.Projects(context.Background())
	if len(list) != 1 || list[0].Name != "Kept" {
		t.Fatalf("projects after reopen = %+v", list)
	}
}
