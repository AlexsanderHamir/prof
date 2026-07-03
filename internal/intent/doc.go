// Package intent owns translation from guided-flow inputs (prof ui, prof tui) into
// github.com/AlexsanderHamir/prof/internal/app.Services calls.
//
// # Supported workflows (catalog)
//
// Each workflow has a stable Kind constant (see kind.go), types in a dedicated file, and the same lifecycle:
// optional Normalize on the intent type, then Validate and Run satisfying Executable.
//
// Supported translations:
//
//   - KindCollect / CollectIntent → Collect.RunAuto
//   - KindCompare / CompareIntent → Tracker.RunTrackAuto (tag layout; not manual file paths)
//   - KindSetup / SetupIntent → Config.CreateDefaultFile (deprecated alias)
//   - KindConfigCreate / ConfigCreateIntent → Config.CreateDefaultFile
//   - KindConfigSave / ConfigSaveIntent → Config.Save
//
// New workflows: add a Kind constant, an entry to AllKinds, a new file with types implementing Executable,
// and wire cli or cli/tui to construct the intent after Survey prompts.
//
// Tests: each intent should have Validate tests and Run tests with fake Services fields.
package intent
