package config

// ResolveCollectionFilter returns the merged function filter for a collection target.
// Precedence: named benchmark/manual entry field overrides defaults; empty fields inherit.
func ResolveCollectionFilter(cfg *Config, target CollectionTarget) FunctionFilter {
	if cfg == nil {
		return FunctionFilter{}
	}

	var named FunctionFilter
	switch {
	case target.auto != "":
		if cfg.Collection.Benchmarks != nil {
			named = cfg.Collection.Benchmarks[target.auto]
		}
	case target.manual != "":
		if cfg.Collection.ManualProfiles != nil {
			named = cfg.Collection.ManualProfiles[target.manual]
		}
	}

	return mergeFunctionFilter(cfg.Collection.Defaults, named)
}

func mergeFunctionFilter(defaults, named FunctionFilter) FunctionFilter {
	out := defaults
	if len(named.IncludePrefixes) > 0 {
		out.IncludePrefixes = named.IncludePrefixes
	}
	if len(named.IgnoreFunctions) > 0 {
		out.IgnoreFunctions = named.IgnoreFunctions
	}
	return out
}
