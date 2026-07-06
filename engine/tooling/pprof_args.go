package tooling

// goToolPprofPrefix returns argv prefix {"go","tool","pprof"}.
func goToolPprofPrefix() []string {
	return []string{"go", "tool", "pprof"}
}

// PprofTextTopArgs returns argv for: go tool pprof -cum -edgefraction=0 -nodefraction=0 -top <binaryPath>
func PprofTextTopArgs(binaryPath string) []string {
	return append(goToolPprofPrefix(),
		"-cum", "-edgefraction=0", "-nodefraction=0", "-top",
		binaryPath,
	)
}

// PprofTextTreeArgs returns argv for: go tool pprof -cum -edgefraction=0 -nodefraction=0 -tree <binaryPath>
func PprofTextTreeArgs(binaryPath string) []string {
	return append(goToolPprofPrefix(),
		"-cum", "-edgefraction=0", "-nodefraction=0", "-tree",
		binaryPath,
	)
}

// PprofPNGArgs returns argv for: go tool pprof -png <binaryPath>
func PprofPNGArgs(binaryPath string) []string {
	return append(append(goToolPprofPrefix(), "-png"), binaryPath)
}

// PprofListArgs returns argv for: go tool pprof -list=<pattern> <binaryPath>
func PprofListArgs(binaryPath, pattern string) []string {
	return append(goToolPprofPrefix(), "-list="+pattern, binaryPath)
}
