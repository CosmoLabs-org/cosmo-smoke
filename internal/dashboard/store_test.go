package dashboard

import (
	"encoding/json"
	"testing"
)

func testStore(t testing.TB, maxRuns ...int) *Store {
	t.Helper()
	mr := 100
	if len(maxRuns) > 0 {
		mr = maxRuns[0]
	}
	s, err := NewStore(":memory:", mr)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func makePayload(project string, total, passed, failed int, durationMs int64) []byte {
	data := map[string]interface{}{
		"project":     project,
		"total":       total,
		"passed":      passed,
		"failed":      failed,
		"skipped":     0,
		"duration_ms": durationMs,
	}
	b, _ := json.Marshal(data)
	return b
}

func TestStore_InsertAndGetProjects(t *testing.T) {
	s := testStore(t)

	s.InsertRun("cosmo-api", makePayload("cosmo-api", 10, 10, 0, 3400))
	s.InsertRun("cosmo-web", makePayload("cosmo-web", 8, 6, 2, 2100))

	projects, err := s.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects = %d, want 2", len(projects))
	}

	api := projects[0]
	if api.Name != "cosmo-api" {
		t.Errorf("project name = %q, want cosmo-api", api.Name)
	}
	if api.LatestStatus != "healthy" {
		t.Errorf("status = %q, want healthy", api.LatestStatus)
	}
	if api.TotalTests != 10 {
		t.Errorf("total = %d, want 10", api.TotalTests)
	}

	web := projects[1]
	if web.LatestStatus != "failing" {
		t.Errorf("status = %q, want failing", web.LatestStatus)
	}
}

func TestStore_GetProjectHistory(t *testing.T) {
	s := testStore(t)

	for i := 0; i < 3; i++ {
		s.InsertRun("cosmo-api", makePayload("cosmo-api", 10, 10, 0, 1000))
	}

	runs, err := s.GetProjectHistory("cosmo-api", 10)
	if err != nil {
		t.Fatalf("GetProjectHistory: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("runs = %d, want 3", len(runs))
	}
	if runs[0].Project != "cosmo-api" {
		t.Errorf("project = %q, want cosmo-api", runs[0].Project)
	}
	if runs[0].Timestamp.Before(runs[2].Timestamp) {
		t.Error("expected descending order by timestamp")
	}
}

func TestStore_HistoryLimit(t *testing.T) {
	s := testStore(t)

	for i := 0; i < 10; i++ {
		s.InsertRun("cosmo-api", makePayload("cosmo-api", 10, 10, 0, 1000))
	}

	runs, err := s.GetProjectHistory("cosmo-api", 5)
	if err != nil {
		t.Fatalf("GetProjectHistory: %v", err)
	}
	if len(runs) != 5 {
		t.Errorf("runs = %d, want 5 (limit)", len(runs))
	}
}

func TestStore_PruneOldRuns(t *testing.T) {
	s := testStore(t, 3)

	for i := 0; i < 5; i++ {
		s.InsertRun("cosmo-api", makePayload("cosmo-api", 10, 10, 0, 1000))
	}

	runs, _ := s.GetProjectHistory("cosmo-api", 100)
	if len(runs) != 3 {
		t.Errorf("runs after prune = %d, want 3", len(runs))
	}
}

func TestStore_EmptyProjects(t *testing.T) {
	s := testStore(t)

	projects, err := s.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("projects = %d, want 0", len(projects))
	}
}

func TestStore_NonexistentProject(t *testing.T) {
	s := testStore(t)

	runs, err := s.GetProjectHistory("nonexistent", 10)
	if err != nil {
		t.Fatalf("GetProjectHistory: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("runs = %d, want 0", len(runs))
	}
}

func TestStore_InvalidPayload(t *testing.T) {
	s := testStore(t)

	_, err := s.InsertRun("bad", []byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON payload")
	}
}

func TestStore_AgeSeconds(t *testing.T) {
	s := testStore(t)
	s.InsertRun("cosmo-api", makePayload("cosmo-api", 5, 5, 0, 500))

	projects, _ := s.GetProjects()
	if len(projects) != 1 {
		t.Fatal("expected 1 project")
	}
	if projects[0].LastRunAgeSeconds > 2 {
		t.Errorf("age = %d, expected < 2", projects[0].LastRunAgeSeconds)
	}
}
