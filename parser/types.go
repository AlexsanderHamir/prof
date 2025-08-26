package parser

type LineObj struct {
	FnName         string
	Flat           float64
	FlatPercentage float64
	SumPercentage  float64
	Cum            float64
	CumPercentage  float64
}

// PackageGroup represents a group of functions from the same package
type PackageGroup struct {
	Name           string
	Functions      []*FunctionInfo
	TotalFlat      float64
	TotalCum       float64
	FlatPercentage float64
	CumPercentage  float64
}

// FunctionInfo represents a function with its performance metrics
type FunctionInfo struct {
	Name           string
	FullName       string
	Flat           float64
	FlatPercentage float64
	Cum            float64
	CumPercentage  float64
	SumPercentage  float64
}
