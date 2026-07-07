package collect

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/datamap"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

type emitMapParams struct {
	Tag              string
	Benchmark        string
	Profiles         []string
	Filter           config.FunctionFilter
	BenchCount       int
	CollectionMode   string
	PerProfile       []datamap.ProfileSnapshot
	IncludeMeasuring bool
}

func emitBenchmarkMap(session *termui.Session, layout workspace.TagLayout, params emitMapParams) {
	pkg := ""
	if params.CollectionMode == datamapCollectionAuto {
		pkg = benchmarkImportPath(params.Benchmark)
	}

	m, err := datamap.Build(datamap.BuildInput{
		Layout:           layout,
		Tag:              params.Tag,
		Benchmark:        params.Benchmark,
		Package:          pkg,
		CollectionMode:   params.CollectionMode,
		Profiles:         params.Profiles,
		Filter:           params.Filter,
		BenchCount:       params.BenchCount,
		PerProfile:       params.PerProfile,
		IncludeMeasuring: params.IncludeMeasuring,
	})
	if err != nil {
		warnMapEmit(session, fmt.Sprintf("benchmark map build failed for %s: %v", params.Benchmark, err))
		return
	}
	path := layout.DataMapping(params.Benchmark)
	if writeErr := datamap.WriteJSON(path, m); writeErr != nil {
		warnMapEmit(session, fmt.Sprintf("benchmark map write failed for %s: %v", params.Benchmark, writeErr))
		return
	}
	slog.Info("Wrote benchmark map", "path", path, "benchmark", params.Benchmark)
}

func warnMapEmit(session *termui.Session, msg string) {
	if session != nil && session.Interactive() {
		session.Warn(msg)
		return
	}
	slog.Warn(msg)
}

const (
	datamapCollectionAuto   = "auto"
	datamapCollectionManual = "manual"
)

func benchmarkImportPath(benchmarkName string) string {
	moduleRoot, err := workspace.FindModuleRoot()
	if err != nil {
		return ""
	}
	pkgDir, err := findBenchmarkPackageDir(moduleRoot, benchmarkName)
	if err != nil {
		return ""
	}
	modPath := readModulePath(filepath.Join(moduleRoot, "go.mod"))
	if modPath == "" {
		return ""
	}
	rel, err := filepath.Rel(moduleRoot, pkgDir)
	if err != nil {
		return modPath
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return modPath
	}
	return modPath + "/" + rel
}

func readModulePath(goModPath string) string {
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
