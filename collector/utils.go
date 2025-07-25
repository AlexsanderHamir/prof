package collector

func getPprofTextParams() []string {
	return []string{
		"-cum",
		"-edgefraction=0",
		"-nodefraction=0",
		"-top",
	}
}
