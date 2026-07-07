package datamap

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var benchLineRE = regexp.MustCompile(`^\S+\s+\d+\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)

const benchResultPass = "PASS"

func parseMeasurementSummary(path string) (*MeasurementSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var nsValues []int64
	var bytesPerOp, allocsPerOp int64
	result := ""
	var elapsed float64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ok\t") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if d, parseErr := strconv.ParseFloat(strings.TrimSuffix(fields[len(fields)-1], "s"), 64); parseErr == nil {
					elapsed = d
				}
			}
		}
		if line == benchResultPass {
			result = benchResultPass
		}
		m := benchLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		ns, _ := strconv.ParseInt(m[1], 10, 64)
		nsValues = append(nsValues, ns)
		bytesPerOp, _ = strconv.ParseInt(m[2], 10, 64)
		allocsPerOp, _ = strconv.ParseInt(m[3], 10, 64)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if len(nsValues) == 0 {
		return nil, fmt.Errorf("no benchmark lines in %s", path)
	}
	sort.Slice(nsValues, func(i, j int) bool { return nsValues[i] < nsValues[j] })
	median := nsValues[len(nsValues)/2]
	return &MeasurementSummary{
		Count:          len(nsValues),
		NsPerOpMedian:  median,
		BytesPerOp:     bytesPerOp,
		AllocsPerOp:    allocsPerOp,
		ElapsedSeconds: elapsed,
		Result:         result,
	}, nil
}
