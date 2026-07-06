package parser

import (
	"sort"

	pprofprofile "github.com/google/pprof/profile"
)

type edgeKey struct {
	caller string
	callee string
}

// BuildCallGraphFromProfile aggregates all nodes and caller→callee edges at valueIndex.
func BuildCallGraphFromProfile(p *pprofprofile.Profile, valueIndex int) *CallGraphData {
	flat, cum := flatAndCumulativeFromSamples(p, valueIndex)
	total := totalSampleValue(p, valueIndex)
	edges := edgesFromSamples(p, valueIndex)

	names := make(map[string]struct{}, len(flat)+len(cum))
	for fn := range flat {
		names[fn] = struct{}{}
	}
	for fn := range cum {
		names[fn] = struct{}{}
	}

	nodes := make([]CallGraphNode, 0, len(names))
	for fn := range names {
		flatVal := flat[fn]
		var flatPct, cumPct float64
		if total > 0 {
			flatPct = float64(flatVal) / float64(total) * 100
			cumPct = float64(cum[fn]) / float64(total) * 100
		}
		nodes = append(nodes, CallGraphNode{
			Name:    fn,
			Flat:    flatVal,
			FlatPct: flatPct,
			Cum:     cum[fn],
			CumPct:  cumPct,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Flat != nodes[j].Flat {
			return nodes[i].Flat > nodes[j].Flat
		}
		return nodes[i].Name < nodes[j].Name
	})

	edgeList := make([]CallGraphEdge, 0, len(edges))
	for k, weight := range edges {
		edgeList = append(edgeList, CallGraphEdge{
			Caller: k.caller,
			Callee: k.callee,
			Weight: weight,
		})
	}
	sort.Slice(edgeList, func(i, j int) bool {
		if edgeList[i].Weight != edgeList[j].Weight {
			return edgeList[i].Weight > edgeList[j].Weight
		}
		if edgeList[i].Caller != edgeList[j].Caller {
			return edgeList[i].Caller < edgeList[j].Caller
		}
		return edgeList[i].Callee < edgeList[j].Callee
	})

	return &CallGraphData{
		Total: total,
		Nodes: nodes,
		Edges: edgeList,
	}
}

func edgesFromSamples(p *pprofprofile.Profile, valueIndex int) map[edgeKey]int64 {
	edges := make(map[edgeKey]int64)
	for _, s := range p.Sample {
		if s == nil || len(s.Value) <= valueIndex {
			continue
		}
		value := s.Value[valueIndex]
		if value == 0 {
			continue
		}
		names := sampleStackNames(s)
		for i := 0; i+1 < len(names); i++ {
			callee := names[i]
			caller := names[i+1]
			k := edgeKey{caller: caller, callee: callee}
			edges[k] += value
		}
	}
	return edges
}

func sampleStackNames(s *pprofprofile.Sample) []string {
	var names []string
	for _, loc := range s.Location {
		if name := locationFunctionName(loc); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func locationFunctionName(loc *pprofprofile.Location) string {
	if loc == nil {
		return ""
	}
	for _, line := range loc.Line {
		if line.Function != nil && line.Function.Name != "" {
			return line.Function.Name
		}
	}
	return ""
}
