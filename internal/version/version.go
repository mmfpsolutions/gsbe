/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

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
