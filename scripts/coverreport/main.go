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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type pkgStat struct {
	sumPct float64
	count  int
}

func main() {
	htmlOut := flag.String("html", "", "write HTML report to this path (via go tool cover -html)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-html coverage.html] coverage.out\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	profile := flag.Arg(0)
	if _, err := os.Stat(profile); err != nil {
		fmt.Fprintf(os.Stderr, "coverreport: %v\n", err)
		os.Exit(1)
	}

	if *htmlOut != "" {
		cmd := exec.Command("go", "tool", "cover", "-html="+profile, "-o", *htmlOut)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "coverreport: html: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "HTML report: %s\n", *htmlOut)
	}

	funcOut, err := exec.Command("go", "tool", "cover", "-func="+profile).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverreport: go tool cover -func: %v\n", err)
		os.Exit(1)
	}

	total, byPkg := parseFuncOutput(funcOut)
	if total < 0 {
		fmt.Fprintf(os.Stderr, "coverreport: could not parse total coverage\n")
		os.Exit(1)
	}

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
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
				if v, err := strconv.ParseFloat(pctStr, 64); err == nil {
					total = v
				}
			}
			continue
		}
		// github.com/.../file.go:42:	FuncName	85.7%
		colon := strings.Index(line, ":")
		if colon <= 0 {
			continue
		}
		filePart := line[:colon]
		if !strings.Contains(filePart, "/") {
			continue
		}
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 2 {
			continue
		}
		pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil {
			continue
		}
		pkg := filepath.ToSlash(filePart)
		if i := strings.LastIndex(pkg, "/"); i >= 0 {
			pkg = pkg[:i]
		}
		st := byPkg[pkg]
		if st == nil {
			st = &pkgStat{}
			byPkg[pkg] = st
		}
		st.sumPct += pct
		st.count++
	}
	return total, byPkg
}
