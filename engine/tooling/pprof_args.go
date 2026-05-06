package tooling

import "github.com/AlexsanderHamir/prof/internal"

// PprofTextTopArgs returns argv for: go tool pprof -cum -edgefraction=0 -nodefraction=0 -top <binaryPath>
func PprofTextTopArgs(binaryPath string) []string {
	return append(internal.GoToolPprofPrefix(),
		"-cum", "-edgefraction=0", "-nodefraction=0", "-top",
		binaryPath,
	)
}

// PprofPNGArgs returns argv for: go tool pprof -png <binaryPath>
func PprofPNGArgs(binaryPath string) []string {
	return append(append(internal.GoToolPprofPrefix(), "-png"), binaryPath)
}

// PprofListArgs returns argv for: go tool pprof -list=<pattern> <binaryPath>
func PprofListArgs(binaryPath, pattern string) []string {
	return append(internal.GoToolPprofPrefix(), "-list="+pattern, binaryPath)
}

// PprofCallgrindArgs returns argv for: go tool pprof -callgrind <binaryPath>
func PprofCallgrindArgs(binaryPath string) []string {
	return append(append(internal.GoToolPprofPrefix(), "-callgrind"), binaryPath)
}
