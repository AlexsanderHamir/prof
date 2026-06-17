package config

// ResolveFilter returns the function filter for a benchmark or manual profile stem.
func ResolveFilter(cfg *Config, name string) FunctionFilter {
	if cfg == nil {
		return FunctionFilter{}
	}
	if global, ok := cfg.FunctionFilter[GlobalSign]; ok {
		return global
	}
	if local, ok := cfg.FunctionFilter[name]; ok {
		return local
	}
	return FunctionFilter{}
}
