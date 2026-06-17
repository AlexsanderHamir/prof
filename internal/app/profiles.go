package app

import "github.com/AlexsanderHamir/prof/engine/collect"

// KnownProfileIDs returns supported profile kind names for discovery and validation.
func KnownProfileIDs() []string {
	return collect.SupportedProfiles
}

// IsKnownProfile reports whether name is a supported profile kind.
func IsKnownProfile(name string) bool {
	for _, id := range KnownProfileIDs() {
		if id == name {
			return true
		}
	}
	return false
}
