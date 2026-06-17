package config

import "strings"

// Normalize cleans and canonicalizes cfg in place.
func Normalize(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}

	cfg.Collection.Defaults = normalizeFunctionFilter(cfg.Collection.Defaults)
	cfg.Collection.Benchmarks = normalizeFunctionFilterMap(cfg.Collection.Benchmarks)
	cfg.Collection.ManualProfiles = normalizeFunctionFilterMap(cfg.Collection.ManualProfiles)

	cfg.Track.Defaults = normalizeTrackPolicy(cfg.Track.Defaults)
	cfg.Track.Benchmarks = normalizeTrackPolicyMap(cfg.Track.Benchmarks)
}

func collectionEmpty(c Collection) bool {
	return functionFilterEmpty(c.Defaults) && c.Benchmarks == nil && c.ManualProfiles == nil
}

func trackSectionEmpty(t Track) bool {
	return trackPolicyEmpty(t.Defaults) && t.Benchmarks == nil
}

func normalizeFunctionFilterMap(m map[string]FunctionFilter) map[string]FunctionFilter {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]FunctionFilter, len(m))
	for k, v := range m {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		v = normalizeFunctionFilter(v)
		if functionFilterEmpty(v) {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeFunctionFilter(f FunctionFilter) FunctionFilter {
	return FunctionFilter{
		IncludePrefixes: dedupeStrings(trimStrings(f.IncludePrefixes)),
		IgnoreFunctions: dedupeStrings(trimStrings(f.IgnoreFunctions)),
	}
}

func functionFilterEmpty(f FunctionFilter) bool {
	return len(f.IncludePrefixes) == 0 && len(f.IgnoreFunctions) == 0
}

func normalizeTrackPolicyMap(m map[string]TrackPolicy) map[string]TrackPolicy {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]TrackPolicy, len(m))
	for k, v := range m {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		v = normalizeTrackPolicy(v)
		if trackPolicyEmpty(v) {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeTrackPolicy(p TrackPolicy) TrackPolicy {
	return TrackPolicy{
		IgnoreFunctions:      dedupeStrings(trimStrings(p.IgnoreFunctions)),
		IgnorePrefixes:       dedupeStrings(trimStrings(p.IgnorePrefixes)),
		MinChangePercent:     p.MinChangePercent,
		MaxRegressionPercent: p.MaxRegressionPercent,
		FailOnImprovement:    p.FailOnImprovement,
	}
}

func trackPolicyEmpty(p TrackPolicy) bool {
	return len(p.IgnoreFunctions) == 0 &&
		len(p.IgnorePrefixes) == 0 &&
		p.MinChangePercent == 0 &&
		p.MaxRegressionPercent == 0 &&
		!p.FailOnImprovement
}

func trimStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
