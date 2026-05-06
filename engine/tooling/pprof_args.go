package tooling

// PprofTextTopArgs returns argv for: go tool pprof -cum -edgefraction=0 -nodefraction=0 -top <binaryPath>
func PprofTextTopArgs(binaryPath string) []string {
	return []string{
		"go", "tool", "pprof",
		"-cum", "-edgefraction=0", "-nodefraction=0", "-top",
		binaryPath,
	}
}

// PprofPNGArgs returns argv for: go tool pprof -png <binaryPath>
func PprofPNGArgs(binaryPath string) []string {
	return []string{"go", "tool", "pprof", "-png", binaryPath}
}

// PprofListArgs returns argv for: go tool pprof -list=<pattern> <binaryPath>
func PprofListArgs(binaryPath, pattern string) []string {
	return []string{"go", "tool", "pprof", "-list=" + pattern, binaryPath}
}

// PprofCallgrindArgs returns argv for: go tool pprof -callgrind <binaryPath>
func PprofCallgrindArgs(binaryPath string) []string {
	return []string{"go", "tool", "pprof", "-callgrind", binaryPath}
}
