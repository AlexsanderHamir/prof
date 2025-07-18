package cli

import (
	"errors"
)

var (
	ErrEmptyBenchmarks = errors.New("benchmarks argument cannot be an empty list")
	ErrEmptyProfiles   = errors.New("profiles argument cannot be an empty list")
	ErrBracket         = errors.New("argument must be wrapped in brackets")
)
