package config

// CollectionTarget identifies which collection filter map entry to resolve.
type CollectionTarget struct {
	auto   string
	manual string
}

// CollectionTargetAuto resolves filters for prof auto benchmark names.
func CollectionTargetAuto(benchmarkName string) CollectionTarget {
	return CollectionTarget{auto: benchmarkName}
}

// CollectionTargetManual resolves filters for prof manual profile file stems.
func CollectionTargetManual(stem string) CollectionTarget {
	return CollectionTarget{manual: stem}
}
