package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// CallGraphFromPath parses path and returns aggregated call-graph data.
func CallGraphFromPath(path string) (*CallGraphData, error) {
	bundle, err := BundleFromPath(path)
	if err != nil {
		return nil, err
	}
	return bundle.CallGraph, nil
}

// WriteCallGraphJSON encodes data as indented JSON at path.
func WriteCallGraphJSON(path string, data *CallGraphData) error {
	if data == nil {
		return errors.New("nil call graph data")
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal call graph json: %w", err)
	}
	if mkdirErr := os.MkdirAll(filepath.Dir(path), workspace.PermDir); mkdirErr != nil {
		return fmt.Errorf("mkdir call graph json parent: %w", mkdirErr)
	}
	if writeErr := os.WriteFile(path, encoded, workspace.PermFile); writeErr != nil {
		return fmt.Errorf("write call graph json: %w", writeErr)
	}
	return nil
}
