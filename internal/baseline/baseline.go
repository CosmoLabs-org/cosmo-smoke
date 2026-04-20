package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

const DefaultFile = ".smoke-baseline.json"

// Entry holds baseline timing for a single test.
type Entry struct {
	DurationMs int64     `json:"duration_ms"`
	Timestamp  time.Time `json:"timestamp"`
}

// File is the on-disk baseline format: test name → timing entry.
type File map[string]Entry

// Comparison holds the result of comparing a test against its baseline.
type Comparison struct {
	Name        string
	BaselineMs  int64
	CurrentMs   int64
	DeltaMs     int64
	DeltaPct    float64
	Regressed   bool
	NewTest     bool
	BaselineNew bool // test existed in baseline but has no previous data
}

// Load reads baseline from disk. Returns empty File if not found.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return nil, fmt.Errorf("reading baseline: %w", err)
	}
	if len(data) == 0 {
		return File{}, nil
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing baseline: %w", err)
	}
	return f, nil
}

// Save writes baseline to disk.
func (f File) Save(path string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Compare checks current test durations against baseline.
// thresholdPct is the regression threshold (e.g. 50 means flag if >150% of baseline).
func (f File) Compare(current map[string]int64, thresholdPct float64) []Comparison {
	var results []Comparison
	for name, curMs := range current {
		entry, exists := f[name]
		if !exists {
			results = append(results, Comparison{
				Name:      name,
				CurrentMs: curMs,
				NewTest:   true,
			})
			continue
		}
		deltaMs := curMs - entry.DurationMs
		deltaPct := 0.0
		if entry.DurationMs > 0 {
			deltaPct = float64(deltaMs) / float64(entry.DurationMs) * 100
		}
		regressed := deltaPct > thresholdPct
		results = append(results, Comparison{
			Name:       name,
			BaselineMs: entry.DurationMs,
			CurrentMs:  curMs,
			DeltaMs:    deltaMs,
			DeltaPct:   deltaPct,
			Regressed:  regressed,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// Update merges current timings into the baseline file.
func (f File) Update(current map[string]int64) {
	now := time.Now().UTC()
	for name, ms := range current {
		f[name] = Entry{DurationMs: ms, Timestamp: now}
	}
}
