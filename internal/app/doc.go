// Package app defines the CLI composition root ([Services]): inject interfaces to swap collect,
// agent, or config backends without changing cobra command wiring.
//
// Use [Default] for production wiring; copy the returned struct and replace individual fields for tests
// or alternate backends. [Services.WithDefaults] fills any nil field from [Default].
package app
