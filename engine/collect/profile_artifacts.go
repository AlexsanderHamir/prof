package collect

import (
	"fmt"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// FailurePolicy controls whether a profile artifact failure fails collection.
type FailurePolicy int

const (
	// Required artifacts fail profile processing on error.
	Required FailurePolicy = iota
	// BestEffort artifacts log a warning and continue.
	BestEffort
)

const (
	artifactHotspots     = "hotspots"
	artifactCallTreeText = "call_tree_text"
	artifactCallGraphPNG = "call_graph_png"
)

// ProduceContext carries inputs for one profile artifact producer.
type ProduceContext struct {
	Runner  tooling.Runner
	Layout  workspace.TagLayout
	Bench   string
	Profile string
	BinPath string
	Session *termui.Session
}

// ArtifactPath resolves the on-disk path for one profile artifact.
type ArtifactPath func(layout workspace.TagLayout, bench, profile string) string

// ProfileArtifact describes one derived output from a profile binary.
type ProfileArtifact struct {
	ID      string
	Policy  FailurePolicy
	Path    ArtifactPath
	Produce func(ProduceContext) error
}

func profileArtifacts() []ProfileArtifact {
	return []ProfileArtifact{
		{
			ID:     artifactHotspots,
			Policy: Required,
			Path:   workspace.TagLayout.Hotspot,
			Produce: func(ctx ProduceContext) error {
				out := ctx.Layout.Hotspot(ctx.Bench, ctx.Profile)
				return runPprofReport(ctx.Runner, tooling.PprofTextReportArgs("top", ctx.BinPath), out)
			},
		},
		{
			ID:     artifactCallTreeText,
			Policy: Required,
			Path:   workspace.TagLayout.CallTreeText,
			Produce: func(ctx ProduceContext) error {
				out := ctx.Layout.CallTreeText(ctx.Bench, ctx.Profile)
				return runPprofReport(ctx.Runner, tooling.PprofTextReportArgs("tree", ctx.BinPath), out)
			},
		},
		{
			ID:     artifactCallGraphPNG,
			Policy: BestEffort,
			Path:   workspace.TagLayout.CallGraph,
			Produce: func(ctx ProduceContext) error {
				return getPNGOutput(ctx.Runner, ctx.BinPath, ctx.Layout.CallGraph(ctx.Profile, ctx.Bench))
			},
		},
	}
}

func emitProfileArtifactsFromCatalog(ctx ProduceContext) error {
	for _, art := range profileArtifacts() {
		if err := art.Produce(ctx); err != nil {
			if art.Policy == BestEffort {
				if art.ID == artifactCallGraphPNG {
					warnSkippedPNG(ctx.Session, ctx.Profile, ctx.Bench, err)
				}
				continue
			}
			return fmt.Errorf("%s: %w", art.ID, err)
		}
	}
	return nil
}

func emitParsedProfileArtifacts(runner tooling.Runner, binPath string, layout workspace.TagLayout, bench, profile string, session *termui.Session) error {
	return emitProfileArtifactsFromCatalog(ProduceContext{
		Runner:  runner,
		Layout:  layout,
		Bench:   bench,
		Profile: profile,
		BinPath: binPath,
		Session: session,
	})
}
