package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// CallGraphFromPath parses path and returns aggregated call-graph data.
func CallGraphFromPath(path string) (*CallGraphData, error) {
	p, err := ParseProfileFromPath(path)
	if err != nil {
		return nil, err
	}
	if err = ValidateProfile(p); err != nil {
		return nil, err
	}
	idx, err := PrimarySampleValueIndex(p)
	if err != nil {
		return nil, err
	}
	if err = ValidateSamplesHaveValueAt(p, idx); err != nil {
		return nil, err
	}
	return BuildCallGraphFromProfile(p, idx), nil
}

// WriteCallGraphJSON encodes data as indented JSON at path.
func WriteCallGraphJSON(path string, data *CallGraphData) error {
	if data == nil {
		return fmt.Errorf("nil call graph data")
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal call graph json: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), workspace.PermDir); err != nil {
		return fmt.Errorf("mkdir call graph json parent: %w", err)
	}
	if err := os.WriteFile(path, encoded, workspace.PermFile); err != nil {
		return fmt.Errorf("write call graph json: %w", err)
	}
	return nil
}
