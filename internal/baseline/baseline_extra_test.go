package baseline

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConcurrentSaveAndLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	f := File{
		"concurrent_test": {DurationMs: 50, Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	const writers = 5
	const readers = 5
	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for i := 0; i < writers; i++ {
		go func(i int) {
			defer wg.Done()
			cp := File{}
			for k, v := range f {
				cp[k] = Entry{DurationMs: v.DurationMs + int64(i), Timestamp: v.Timestamp}
			}
			if err := cp.Save(path); err != nil {
				t.Errorf("save %d: %v", i, err)
			}
		}(i)
	}

	for i := 0; i < readers; i++ {
		go func(i int) {
			defer wg.Done()
			loaded, err := Load(path)
			if err != nil {
				t.Errorf("load %d: %v", i, err)
				return
			}
			if len(loaded) == 0 {
				t.Logf("load %d: empty (file may not exist yet)", i)
			}
		}(i)
	}

	wg.Wait()

	// Final load should succeed with valid data.
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("final load: %v", err)
	}
	if len(loaded) == 0 {
		t.Fatal("expected at least one entry after concurrent writes")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	if err := os.WriteFile(path, []byte("{{not valid json!!}}}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error loading corrupt JSON")
	}
}

func TestNegativeDuration_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	f := File{
		"neg_test": {DurationMs: -100, Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	if err := f.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded["neg_test"].DurationMs != -100 {
		t.Errorf("duration = %d, want -100", loaded["neg_test"].DurationMs)
	}
}

func TestSave_MissingDirectory(t *testing.T) {
	f := File{
		"test": {DurationMs: 10, Timestamp: time.Now()},
	}
	err := f.Save("/nonexistent/deeply/nested/dir/baseline.json")
	if err == nil {
		t.Fatal("expected error saving to missing directory")
	}
}
