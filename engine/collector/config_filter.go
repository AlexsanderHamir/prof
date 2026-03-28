package collector

import "github.com/AlexsanderHamir/prof/internal"

func globalFilterFromConfig(cfg *internal.Config) (internal.FunctionFilter, bool) {
	f, ok := cfg.FunctionFilter[internal.GlobalSign]
	return f, ok
}

func resolveFunctionFilter(cfg *internal.Config, fileName string, global internal.FunctionFilter) internal.FunctionFilter {
	if _, hasGlobal := cfg.FunctionFilter[internal.GlobalSign]; hasGlobal {
		return global
	}
	if local, ok := cfg.FunctionFilter[fileName]; ok {
		return local
	}
	return internal.FunctionFilter{}
}
