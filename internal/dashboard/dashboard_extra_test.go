package dashboard

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStore_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	// Enable WAL mode (persists across connections) and set busy_timeout
	// on the same connection pool that NewStore will use.
	pragmaDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open pragma connection: %v", err)
	}
	pragmaDB.Exec("PRAGMA journal_mode=WAL")
	pragmaDB.Close()

	// Open store with busy_timeout in DSN so all pool connections honor it
	dsn := "file:" + dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	s, err := NewStore(dsn, 100)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	const goroutines = 20
	const writesPer = 5

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			for j := range writesPer {
				_, err := s.InsertRun("concurrent-project", makePayload("concurrent-project", 10, 10, 0, 100))
				if err != nil {
					t.Errorf("goroutine %d write %d: %v", i, j, err)
				}
			}
		}()
	}
	wg.Wait()

	runs, err := s.GetProjectHistory("concurrent-project", 1000)
	if err != nil {
		t.Fatalf("GetProjectHistory: %v", err)
	}
	want := goroutines * writesPer
	if len(runs) != want {
		t.Errorf("got %d runs, want %d (no data loss)", len(runs), want)
	}

	projects, err := s.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 1 || projects[0].Name != "concurrent-project" {
		t.Errorf("projects = %+v, want 1 project named concurrent-project", projects)
	}
}

func TestStore_EmptyProjectName(t *testing.T) {
	s := testStore(t)

	_, err := s.InsertRun("", makePayload("", 5, 5, 0, 200))
	if err != nil {
		t.Fatalf("InsertRun with empty project: %v", err)
	}

	projects, err := s.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(projects))
	}
	if projects[0].Name != "" {
		t.Errorf("project name = %q, want empty", projects[0].Name)
	}

	runs, err := s.GetProjectHistory("", 10)
	if err != nil {
		t.Fatalf("GetProjectHistory: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("runs = %d, want 1", len(runs))
	}
}

func TestStore_SpecialCharactersInNames(t *testing.T) {
	s := testStore(t)

	cases := []struct {
		name    string
		project string
	}{
		{"unicode", "project-日本語-测试-🌟"},
		{"quotes", "project-\"quoted\"-it's"},
		{"backslash", "project\\slash/pipe|thing"},
		{"newlines", "project\nwith\nnewlines"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.InsertRun(tc.project, makePayload(tc.project, 3, 3, 0, 100))
			if err != nil {
				t.Fatalf("InsertRun %s: %v", tc.name, err)
			}

			runs, err := s.GetProjectHistory(tc.project, 10)
			if err != nil {
				t.Fatalf("GetProjectHistory: %v", err)
			}
			if len(runs) != 1 {
				t.Fatalf("runs = %d, want 1", len(runs))
			}
			if runs[0].Project != tc.project {
				t.Errorf("project = %q, want %q", runs[0].Project, tc.project)
			}

			var payload map[string]any
			if err := json.Unmarshal([]byte(runs[0].Payload), &payload); err != nil {
				t.Fatalf("unmarshal payload: %v", err)
			}
			if payload["project"] != tc.project {
				t.Errorf("payload project = %v, want %q", payload["project"], tc.project)
			}
		})
	}
}

func TestStore_PaginationSubset(t *testing.T) {
	s := testStore(t)

	const totalRuns = 10
	for i := range totalRuns {
		s.InsertRun("paged-project", makePayload("paged-project", totalRuns, totalRuns, 0, int64(i*100)))
	}

	// Limited subset
	page1, err := s.GetProjectHistory("paged-project", 3)
	if err != nil {
		t.Fatalf("limit 3: %v", err)
	}
	if len(page1) != 3 {
		t.Fatalf("page len = %d, want 3", len(page1))
	}

	// Descending order
	for i := 1; i < len(page1); i++ {
		if page1[i].Timestamp.After(page1[i-1].Timestamp) {
			t.Errorf("page1[%d] timestamp after page1[%d]", i, i-1)
		}
	}

	// Full set
	all, err := s.GetProjectHistory("paged-project", totalRuns)
	if err != nil {
		t.Fatalf("full: %v", err)
	}
	if len(all) != totalRuns {
		t.Errorf("full len = %d, want %d", len(all), totalRuns)
	}

	// Limited subset is prefix of full set
	for i := range page1 {
		if page1[i].ID != all[i].ID {
			t.Errorf("page1[%d].ID = %d, want %d from full set", i, page1[i].ID, all[i].ID)
		}
	}
}
