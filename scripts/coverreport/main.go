// coverreport prints module and per-package statement coverage from a Go cover profile.
//
// Uses the official `go tool cover -func` output for totals, then groups function
// coverage by import path for a per-package table (sorted lowest first).
//
// Usage:
//
//	go run ./scripts/coverreport/main.go coverage.out
//	go run ./scripts/coverreport/main.go -html coverage.html coverage.out
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

type pkgStat struct {
	sumPct float64
	count  int
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	htmlOut := fs.String("html", "", "write HTML report to this path (via go tool cover -html)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-html coverage.html] coverage.out\n", filepath.Base(os.Args[0]))
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}

	profile, err := validatePath(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverreport: %v\n", err)
		return 1
	}

	if *htmlOut != "" {
		outPath, htmlErr := validatePath(*htmlOut)
		if htmlErr != nil {
			fmt.Fprintf(os.Stderr, "coverreport: html output: %v\n", htmlErr)
			return 1
		}
		if err = runCoverHTML(profile, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "coverreport: html: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "HTML report: %s\n", outPath)
	}

	funcOut, err := runCoverFunc(profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverreport: go tool cover -func: %v\n", err)
		return 1
	}

	total, byPkg := parseFuncOutput(funcOut)
	if total < 0 {
		fmt.Fprintf(os.Stderr, "coverreport: could not parse total coverage\n")
		return 1
	}

	printReport(total, byPkg)
	return 0
}

func validatePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if strings.Contains(path, "\x00") {
		return "", errors.New("invalid path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func runCoverHTML(profile, htmlOut string) error {
	_, err := tooling.NewExecRunner().Run(context.Background(),
		[]string{"go", "tool", "cover", "-html=" + profile, "-o", htmlOut},
		tooling.RunOpts{Stdout: os.Stdout, Stderr: os.Stderr})
	return err
}

func runCoverFunc(profile string) ([]byte, error) {
	return tooling.NewExecRunner().Run(context.Background(),
		[]string{"go", "tool", "cover", "-func=" + profile},
		tooling.RunOpts{})
}

func printReport(total float64, byPkg map[string]*pkgStat) {
	names := make([]string, 0, len(byPkg))
	for name, st := range byPkg {
		if st.count == 0 {
			continue
		}
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		pi := avgPct(byPkg[names[i]])
		pj := avgPct(byPkg[names[j]])
		if pi == pj {
			return names[i] < names[j]
		}
		return pi < pj
	})

	fmt.Printf("Total statement coverage: %.1f%%\n", total)
	fmt.Println("(from go tool cover -func; uses -coverpkg=./... profile)")
	fmt.Println()
	fmt.Printf("%-52s %8s %10s\n", "Package", "Cover", "Functions")
	fmt.Printf("%-52s %8s %10s\n", strings.Repeat("-", 52), "------", "----------")
	for _, name := range names {
		st := byPkg[name]
		fmt.Printf("%-52s %7.1f%% %8d\n", name, avgPct(st), st.count)
	}
}

func avgPct(st *pkgStat) float64 {
	if st.count == 0 {
		return 0
	}
	return st.sumPct / float64(st.count)
}

func parseFuncOutput(raw []byte) (total float64, byPkg map[string]*pkgStat) {
	byPkg = make(map[string]*pkgStat)
	total = -1
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "total:") {
			if v, ok := parseTotalLine(line); ok {
				total = v
			}
			continue
		}
		if pkg, pct, ok := parseFuncLine(line); ok {
			recordPkgPct(byPkg, pkg, pct)
		}
	}
	return total, byPkg
}

func parseTotalLine(line string) (float64, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return 0, false
	}
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	v, err := strconv.ParseFloat(pctStr, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseFuncLine(line string) (pkg string, pct float64, ok bool) {
	colon := strings.Index(line, ":")
	if colon <= 0 {
		return "", 0, false
	}
	filePart := line[:colon]
	if !strings.Contains(filePart, "/") {
		return "", 0, false
	}
	fields := strings.Fields(line[colon+1:])
	if len(fields) < 2 {
		return "", 0, false
	}
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	pct, err := strconv.ParseFloat(pctStr, 64)
	if err != nil {
		return "", 0, false
	}
	pkg = filepath.ToSlash(filePart)
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		pkg = pkg[:i]
	}
	return pkg, pct, true
}

func recordPkgPct(byPkg map[string]*pkgStat, pkg string, pct float64) {
	st := byPkg[pkg]
	if st == nil {
		st = &pkgStat{}
		byPkg[pkg] = st
	}
	st.sumPct += pct
	st.count++
}
