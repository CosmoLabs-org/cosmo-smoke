package dashboard

// Config holds dashboard configuration.
type Config struct {
	Enabled          bool
	DBPath           string
	APIKey           string
	MaxRunsPerProject int
}
