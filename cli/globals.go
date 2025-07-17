package cli

import (
	"errors"
)

var (
	ErremptyBenchmarks = errors.New("benchmarks argument cannot be an empty list")
	ErremptyProfiles   = errors.New("profiles argument cannot be an empty list")
	Errbracket         = errors.New("argument must be wrapped in brackets")
)
