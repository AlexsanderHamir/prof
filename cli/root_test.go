package cli

import (
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

func TestCreateRootCmdNilServices(t *testing.T) {
	cmd := CreateRootCmd(nil)
	if cmd == nil || cmd.Use != "prof" {
		t.Fatalf("unexpected root: %#v", cmd)
	}
	if len(cmd.Commands()) == 0 {
		t.Fatal("expected subcommands")
	}
}

func TestCreateRootCmdWithPartialServices(t *testing.T) {
	cmd := CreateRootCmd(&app.Services{})
	if cmd == nil {
		t.Fatal("nil command")
	}
}
