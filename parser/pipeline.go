package parser

import (
	"fmt"
	"io"
	"os"

	pprofprofile "github.com/google/pprof/profile"
)

// ProfileOpener obtains a byte stream for a profile location (local path, URL, object key, etc.).
type ProfileOpener interface {
	Open(path string) (io.ReadCloser, error)
}

// ProfileDecoder turns raw bytes into a pprof profile. Swap to use a different format or parser.
type ProfileDecoder interface {
	Decode(r io.Reader) (*pprofprofile.Profile, error)
}

// ProfileValidator rejects malformed profiles before aggregation.
type ProfileValidator interface {
	Validate(p *pprofprofile.Profile) error
}

// SampleIndexSelector chooses which entry in Sample.Value to aggregate (e.g. first sample type vs wall vs alloc).
type SampleIndexSelector interface {
	PrimaryIndex(p *pprofprofile.Profile) (int, error)
}

// SampleShapeChecker ensures each sample is usable at the chosen value index.
type SampleShapeChecker interface {
	EnsureValueAt(p *pprofprofile.Profile, index int) error
}

// ProfileAggregator builds ProfileData from a parsed profile and value index.
type ProfileAggregator interface {
	Aggregate(p *pprofprofile.Profile, valueIndex int) *ProfileData
}

// Pipeline composes the stages from open → decode → validate → normalize → aggregate.
// Zero values in fields are replaced by stock implementations when you call RunFromReader or RunFromPath.
type Pipeline struct {
	Opener      ProfileOpener
	Decoder     ProfileDecoder
	Validator   ProfileValidator
	IndexSelect SampleIndexSelector
	IndexCheck  SampleShapeChecker
	Aggregator  ProfileAggregator
}

// DefaultPipeline returns an independent Pipeline with standard components; copy and replace fields to customize.
func DefaultPipeline() Pipeline {
	return newStdPipeline()
}

func newStdPipeline() Pipeline {
	return Pipeline{
		Opener:      FileProfileOpener{},
		Decoder:     PProfDecoder{},
		Validator:   StandardProfileValidator{},
		IndexSelect: FirstSampleIndexSelector{},
		IndexCheck:  AllSamplesValueChecker{},
		Aggregator:  FlatCumAggregator{},
	}
}

var stdPipeline = newStdPipeline()

func (pl Pipeline) withDefaults() Pipeline {
	if pl.Opener == nil {
		pl.Opener = FileProfileOpener{}
	}
	if pl.Decoder == nil {
		pl.Decoder = PProfDecoder{}
	}
	if pl.Validator == nil {
		pl.Validator = StandardProfileValidator{}
	}
	if pl.IndexSelect == nil {
		pl.IndexSelect = FirstSampleIndexSelector{}
	}
	if pl.IndexCheck == nil {
		pl.IndexCheck = AllSamplesValueChecker{}
	}
	if pl.Aggregator == nil {
		pl.Aggregator = FlatCumAggregator{}
	}
	return pl
}

// RunFromReader executes decode → validate → index → check → aggregate.
func (pl Pipeline) RunFromReader(r io.Reader) (*ProfileData, error) {
	pl = pl.withDefaults()
	p, err := pl.Decoder.Decode(r)
	if err != nil {
		return nil, err
	}
	if err = pl.Validator.Validate(p); err != nil {
		return nil, err
	}
	idx, err := pl.IndexSelect.PrimaryIndex(p)
	if err != nil {
		return nil, err
	}
	if err = pl.IndexCheck.EnsureValueAt(p, idx); err != nil {
		return nil, err
	}
	return pl.Aggregator.Aggregate(p, idx), nil
}

// RunFromPath executes open → RunFromReader.
func (pl Pipeline) RunFromPath(path string) (*ProfileData, error) {
	pl = pl.withDefaults()
	rc, err := pl.Opener.Open(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return pl.RunFromReader(rc)
}

// --- Default implementations (replace by embedding these and overriding, or implement interfaces yourself) ---

// FileProfileOpener opens a local file path.
type FileProfileOpener struct{}

// Open opens the profile file at path.
func (FileProfileOpener) Open(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	return f, nil
}

// PProfDecoder uses github.com/google/pprof/profile.Parse.
type PProfDecoder struct{}

// Decode parses a pprof profile from r.
func (PProfDecoder) Decode(r io.Reader) (*pprofprofile.Profile, error) {
	p, err := pprofprofile.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pprof profile: %w", err)
	}
	return p, nil
}

// StandardProfileValidator applies CheckValid and sample-type guards.
type StandardProfileValidator struct{}

// Validate runs profile validation and sample-type checks.
func (StandardProfileValidator) Validate(p *pprofprofile.Profile) error {
	return ValidateProfile(p)
}

// FirstSampleIndexSelector uses the last sample value index (pprof -top default).
type FirstSampleIndexSelector struct{}

// PrimaryIndex returns the primary sample value index (last sample type).
func (FirstSampleIndexSelector) PrimaryIndex(p *pprofprofile.Profile) (int, error) {
	return PrimarySampleValueIndex(p)
}

// AllSamplesValueChecker requires every sample to have a value at the chosen index.
type AllSamplesValueChecker struct{}

// EnsureValueAt checks that every sample has a value at index.
func (AllSamplesValueChecker) EnsureValueAt(p *pprofprofile.Profile, index int) error {
	return ValidateSamplesHaveValueAt(p, index)
}

// FlatCumAggregator uses the package flat/cumulative aggregation rules.
type FlatCumAggregator struct{}

// Aggregate builds flat/cumulative profile data at valueIndex.
func (FlatCumAggregator) Aggregate(p *pprofprofile.Profile, valueIndex int) *ProfileData {
	return AggregateProfileData(p, valueIndex)
}
