package cmd

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestIsRelevantEvent(t *testing.T) {
	tests := []struct {
		name string
		op   fsnotify.Op
		want bool
	}{
		{"write", fsnotify.Write, true},
		{"create", fsnotify.Create, true},
		{"rename", fsnotify.Rename, true},
		{"remove", fsnotify.Remove, true},
		{"chmod only", fsnotify.Chmod, false},
		{"write+chmod", fsnotify.Write | fsnotify.Chmod, true},
		{"zero", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRelevantEvent(tt.op); got != tt.want {
				t.Errorf("isRelevantEvent(%v) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}
