package version

import "time"

// Set via ldflags at build time
var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	Commit    = "unknown"
)

// StartTime records when the application started
var StartTime time.Time

func init() {
	StartTime = time.Now()
}

// Uptime returns the duration since the application started
func Uptime() time.Duration {
	return time.Since(StartTime)
}
