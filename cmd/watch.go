package cmd

import "github.com/fsnotify/fsnotify"

// isRelevantEvent returns true if the event warrants a re-run.
func isRelevantEvent(op fsnotify.Op) bool {
	return op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) != 0
}
