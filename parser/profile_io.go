package parser

import (
	"errors"
	"fmt"
	"io"

	pprofprofile "github.com/google/pprof/profile"
)

// LoadProfile opens a pprof file for reading. The caller must close the returned ReadCloser.
func LoadProfile(path string) (io.ReadCloser, error) {
	return stdPipeline.withDefaults().Opener.Open(path)
}

// ParseProfileFromReader decodes a pprof profile from r.
func ParseProfileFromReader(r io.Reader) (*pprofprofile.Profile, error) {
	return stdPipeline.withDefaults().Decoder.Decode(r)
}

// ParseProfileFromPath opens path and parses a pprof profile (Load + ParseProfileFromReader).
func ParseProfileFromPath(path string) (*pprofprofile.Profile, error) {
	rc, err := LoadProfile(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ParseProfileFromReader(rc)
}

// ValidateProfile runs pprof structural checks and guards required for aggregation.
func ValidateProfile(p *pprofprofile.Profile) error {
	if p == nil {
		return errors.New("nil profile")
	}
	if err := p.CheckValid(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}
	if len(p.SampleType) == 0 {
		return errors.New("profile has no sample types")
	}
	return nil
}

// PrimarySampleValueIndex returns the index into Sample.Value used for flat/cum aggregation.
// Uses the last sample type, matching go tool pprof report.NewDefault (pprof -top).
func PrimarySampleValueIndex(p *pprofprofile.Profile) (int, error) {
	if len(p.SampleType) == 0 {
		return 0, errors.New("profile has no sample types")
	}
	return len(p.SampleType) - 1, nil
}

// ValidateSamplesHaveValueAt ensures every sample has a value at the given index.
func ValidateSamplesHaveValueAt(p *pprofprofile.Profile, index int) error {
	if index < 0 {
		return fmt.Errorf("negative sample value index: %d", index)
	}
	for i, s := range p.Sample {
		if s == nil {
			return fmt.Errorf("sample %d is nil", i)
		}
		if len(s.Value) <= index {
			return fmt.Errorf("sample %d: len(Value)=%d, need index %d", i, len(s.Value), index)
		}
	}
	return nil
}

// ProfileDataFromReader runs parse → validate → normalize (sample index) → aggregate using the default pipeline.
func ProfileDataFromReader(r io.Reader) (*ProfileData, error) {
	return stdPipeline.RunFromReader(r)
}

// profileDataFromPath loads a file through the default pipeline (path entry).
func profileDataFromPath(profilePath string) (*ProfileData, error) {
	return stdPipeline.RunFromPath(profilePath)
}
