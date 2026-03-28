package parser

import (
	"fmt"
	"sort"
	"strings"
)

func sortPackagesByFlatPercentage(packageGroups map[string]*PackageGroup) []*PackageGroup {
	packages := make([]*PackageGroup, 0, len(packageGroups))
	for _, pkg := range packageGroups {
		packages = append(packages, pkg)
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].FlatPercentage > packages[j].FlatPercentage
	})
	return packages
}

func formatFunctionLine(fn *FunctionInfo, isUnknownPackage bool) string {
	if isUnknownPackage {
		return fmt.Sprintf("- `%s` → flat: %.2f, flat%%: %.2f%%, sum%%: %.2f%%, cum: %.2f, cum%%: %.2f%%\n",
			fn.Name, fn.Flat, fn.FlatPercentage, fn.SumPercentage, fn.Cum, fn.CumPercentage)
	}
	return fmt.Sprintf("- `%s` → %.2f%%\n", fn.Name, fn.FlatPercentage)
}

func formatPackageReport(packages []*PackageGroup) string {
	var b strings.Builder
	for i, pkg := range packages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(fmt.Sprintf("#### **%s**\n", pkg.Name))
		sort.Slice(pkg.Functions, func(i, j int) bool {
			return pkg.Functions[i].FlatPercentage > pkg.Functions[j].FlatPercentage
		})
		unknown := pkg.Name == "unknown"
		for _, fn := range pkg.Functions {
			b.WriteString(formatFunctionLine(fn, unknown))
		}
		b.WriteString(fmt.Sprintf("\n**Subtotal (%s)**: ≈%.1f%%",
			shortPackageLabel(pkg.Name), pkg.FlatPercentage))
	}
	return b.String()
}
