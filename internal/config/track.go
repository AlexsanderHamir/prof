package config

// ResolveTrackPolicy returns merged track policy for a benchmark name.
// Precedence: track.benchmarks[name] overrides track.defaults field-by-field.
func ResolveTrackPolicy(cfg *Config, benchmarkName string) TrackPolicy {
	if cfg == nil {
		return TrackPolicy{}
	}
	named := TrackPolicy{}
	if cfg.Track.Benchmarks != nil {
		named = cfg.Track.Benchmarks[benchmarkName]
	}
	return mergeTrackPolicy(cfg.Track.Defaults, named)
}

func mergeTrackPolicy(defaults, named TrackPolicy) TrackPolicy {
	out := defaults
	if len(named.IgnoreFunctions) > 0 {
		out.IgnoreFunctions = named.IgnoreFunctions
	}
	if len(named.IgnorePrefixes) > 0 {
		out.IgnorePrefixes = named.IgnorePrefixes
	}
	if named.MinChangePercent > 0 {
		out.MinChangePercent = named.MinChangePercent
	}
	if named.MaxRegressionPercent > 0 {
		out.MaxRegressionPercent = named.MaxRegressionPercent
	}
	if named.FailOnImprovement {
		out.FailOnImprovement = true
	}
	return out
}

// ShouldIgnoreFunction reports whether functionName matches track policy ignores.
func ShouldIgnoreFunction(policy TrackPolicy, functionName string) bool {
	for _, ignored := range policy.IgnoreFunctions {
		if functionName == ignored {
			return true
		}
	}
	for _, prefix := range policy.IgnorePrefixes {
		if len(prefix) > 0 && hasPrefix(functionName, prefix) {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
