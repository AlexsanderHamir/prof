package app

import "github.com/AlexsanderHamir/prof/engine/tracker"

// TrackOutputFormats lists allowed track report formats.
func TrackOutputFormats() []string {
	return tracker.ValidOutputFormats
}

// ValidTrackOutputFormat reports whether format is supported for track commands.
func ValidTrackOutputFormat(format string) bool {
	return tracker.ValidOutputFormat(format)
}
