package tooling

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ProfileKind describes one go test profiling output (pprof binary) the benchmark pipeline supports.
type ProfileKind struct {
	ID          string
	GoTestFlag  string // e.g. -cpuprofile=cpu.out
	OutFileName string // basename written in the package directory before moves (e.g. cpu.out)
}

// Catalog holds supported profile kinds and helpers to build go test / path logic from them.
type Catalog struct {
	profiles []ProfileKind
	byID     map[string]ProfileKind
}

// DefaultCatalog returns the stock profile kinds (cpu, memory, mutex, block).
func DefaultCatalog() *Catalog {
	profiles := []ProfileKind{
		{ID: "cpu", GoTestFlag: "-cpuprofile=cpu.out", OutFileName: "cpu.out"},
		{ID: "memory", GoTestFlag: "-memprofile=memory.out", OutFileName: "memory.out"},
		{ID: "mutex", GoTestFlag: "-mutexprofile=mutex.out", OutFileName: "mutex.out"},
		{ID: "block", GoTestFlag: "-blockprofile=block.out", OutFileName: "block.out"},
	}
	byID := make(map[string]ProfileKind, len(profiles))
	for _, p := range profiles {
		byID[p.ID] = p
	}
	return &Catalog{profiles: profiles, byID: byID}
}

// ProfileIDs returns supported profile identifiers in stable sorted order.
func (c *Catalog) ProfileIDs() []string {
	if c == nil {
		return nil
	}
	ids := make([]string, 0, len(c.profiles))
	for _, p := range c.profiles {
		ids = append(ids, p.ID)
	}
	sort.Strings(ids)
	return ids
}

// ProfileIDsSorted matches historical SupportedProfiles order (declaration order) for UI/tests that depend on it.
func (c *Catalog) ProfileIDsSorted() []string {
	if c == nil {
		return nil
	}
	out := make([]string, len(c.profiles))
	for i, p := range c.profiles {
		out[i] = p.ID
	}
	return out
}

// ValidateProfile returns an error if id is not a supported profile kind.
func (c *Catalog) ValidateProfile(id string) error {
	if c == nil {
		return errors.New("tooling: nil catalog")
	}
	if _, ok := c.byID[id]; !ok {
		return fmt.Errorf("profile %s is not supported", id)
	}
	return nil
}

// GoTestProfileArgs returns go test argv fragments (one flag per id) for the given profile ids in order.
func (c *Catalog) GoTestProfileArgs(ids []string) ([]string, error) {
	if c == nil {
		return nil, errors.New("tooling: nil catalog")
	}
	var out []string
	for _, id := range ids {
		p, ok := c.byID[id]
		if !ok {
			return nil, fmt.Errorf("profile %s is not supported", id)
		}
		out = append(out, p.GoTestFlag)
	}
	return out, nil
}

// OutFileName returns the intermediate output basename go test writes for this profile (before move to bench/).
func (c *Catalog) OutFileName(profileID string) (string, bool) {
	if c == nil {
		return "", false
	}
	p, ok := c.byID[profileID]
	if !ok {
		return "", false
	}
	return p.OutFileName, true
}

// ProfileKinds returns a copy of registered profile kinds in declaration order.
func (c *Catalog) ProfileKinds() []ProfileKind {
	if c == nil {
		return nil
	}
	out := make([]ProfileKind, len(c.profiles))
	copy(out, c.profiles)
	return out
}

// KnownProfileSet returns a set of profile ids for membership checks (for example CLI discovery).
func (c *Catalog) KnownProfileSet() map[string]struct{} {
	if c == nil {
		return nil
	}
	m := make(map[string]struct{}, len(c.byID))
	for id := range c.byID {
		m[id] = struct{}{}
	}
	return m
}

// NormalizeProfileCSV splits comma-separated profile names and trims spaces; empty entries are dropped.
func NormalizeProfileCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
