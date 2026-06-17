package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func newProfCmd(t *testing.T, envFullPath string, args []string) *exec.Cmd {
	t.Helper()

	profBinary := filepath.Join(envFullPath, profBinaryName())
	if _, err := os.Stat(profBinary); os.IsNotExist(err) {
		t.Fatalf("prof binary not found at: %s", profBinary)
	}

	cmd := exec.Command(profBinary, args...)
	cmd.Dir = envFullPath
	return cmd
}

// runProfCaptured runs prof like runProf but returns stdout/stderr for assertions.
func runProfCaptured(t *testing.T, envFullPath string, args []string, expectedErrMessage string, checkSuccessMessage bool) (stdout, stderr string, shouldContinue bool) {
	t.Helper()

	var stdoutB, stderrB bytes.Buffer
	cmd := newProfCmd(t, envFullPath, args)
	cmd.Stdout = &stdoutB
	cmd.Stderr = &stderrB

	err := cmd.Run()
	if err != nil {
		shouldContinue = handleCommandError(t, err, &stdoutB, &stderrB, expectedErrMessage)
		return stdoutB.String(), stderrB.String(), shouldContinue
	}

	if checkSuccessMessage && !strings.Contains(stderrB.String(), workspace.InfoCollectionSuccess) {
		t.Fatal("Expected success message not found")
	}

	return stdoutB.String(), stderrB.String(), true
}

func runProf(t *testing.T, envFullPath string, args []string, expectedErrMessage string, checkSuccessMessage bool) (shouldContinue bool) {
	t.Helper()
	_, _, ok := runProfCaptured(t, envFullPath, args, expectedErrMessage, checkSuccessMessage)
	return ok
}

func handleCommandError(t *testing.T, err error, stdout, stderr *bytes.Buffer, expectedErrMessage string) bool {
	t.Helper()

	if expectedErrMessage != "" {
		stderrText := stderr.String()
		if strings.Contains(stderrText, expectedErrMessage) {
			return false
		}

		t.Fatalf("Expected error message '%s' not found.\nStderr: %s\nStdout: %s",
			expectedErrMessage, stderrText, stdout.String())
	}

	t.Fatalf("prof command failed: %v\nStdout: %s\nStderr: %s",
		err, stdout.String(), stderr.String())

	return true // Never reached in case of t.Fatalf.
}
