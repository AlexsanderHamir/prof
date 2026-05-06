package cursoragent

// Adapted from T2A pkgs/agents/runner/cursor/protocol.go and progress.go (stream-json NDJSON parsing).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	eventSystem    = "system"
	eventAssistant = "assistant"
	eventToolCall  = "tool_call"
	eventResult    = "result"

	subtypeInit      = "init"
	subtypeStarted   = "started"
	subtypeStart     = "start"
	subtypeCompleted = "completed"
	subtypeSuccess   = "success"
	subtypeDone      = "done"
	subtypeFailed    = "failed"
	subtypeError     = "error"

	contentText = "text"
)

type progressMessage struct {
	Role    string            `json:"role,omitempty"`
	Content []progressContent `json:"content,omitempty"`
}

type progressContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type streamEventHead struct {
	Type      string          `json:"type,omitempty"`
	Subtype   string          `json:"subtype,omitempty"`
	Model     string          `json:"model,omitempty"`
	CallID    string          `json:"call_id,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Message   progressMessage `json:"message,omitempty"`
}

type parsedOutput struct {
	Type                  string `json:"type,omitempty"`
	Subtype               string `json:"subtype,omitempty"`
	IsError               bool   `json:"is_error,omitempty"`
	Result                string `json:"result,omitempty"`
	SessionID             string `json:"session_id,omitempty"`
	ResolvedModel         string `json:"-"`
	MissingTerminalResult bool   `json:"-"`
}

func textContent(parts []progressContent) string {
	var b strings.Builder
	for _, part := range parts {
		if part.Type != contentText {
			continue
		}
		text := strings.TrimSpace(part.Text)
		if text == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(text)
	}
	return b.String()
}

// streamAccum tracks NDJSON stream-json state while scanning stdout.
type streamAccum struct {
	out               parsedOutput
	gotResult         bool
	lastDecErr        error
	lastAssistantText string
	lastSessionID     string
	openToolCalls     map[string]struct{}
	openAnonymous     int
}

func trySingleResultJSON(stdout []byte) (parsedOutput, bool) {
	var single parsedOutput
	if err := json.Unmarshal(stdout, &single); err == nil && single.Type == eventResult {
		return single, true
	}
	return parsedOutput{}, false
}

func (a *streamAccum) noteSessionID(head streamEventHead) {
	if a.lastSessionID == "" {
		a.lastSessionID = strings.TrimSpace(head.SessionID)
	}
}

func (a *streamAccum) ingestSystem(head streamEventHead) {
	if head.Subtype == subtypeInit && a.out.ResolvedModel == "" {
		a.out.ResolvedModel = strings.TrimSpace(head.Model)
	}
	a.noteSessionID(head)
}

func (a *streamAccum) ingestAssistant(head streamEventHead) {
	if msg := strings.TrimSpace(textContent(head.Message.Content)); msg != "" {
		a.lastAssistantText = msg
	}
	a.noteSessionID(head)
}

func (a *streamAccum) ingestToolCall(head streamEventHead) {
	updateOpenToolCalls(a.openToolCalls, &a.openAnonymous, head)
	a.noteSessionID(head)
}

func (a *streamAccum) ingestResult(raw []byte) {
	var evt parsedOutput
	if err := json.Unmarshal(raw, &evt); err != nil {
		a.lastDecErr = err
		return
	}
	resolved := a.out.ResolvedModel
	a.out = evt
	a.out.ResolvedModel = resolved
	a.gotResult = true
}

func (a *streamAccum) ingest(raw []byte) {
	var head streamEventHead
	if err := json.Unmarshal(raw, &head); err != nil {
		a.lastDecErr = err
		return
	}
	switch head.Type {
	case eventSystem:
		a.ingestSystem(head)
	case eventAssistant:
		a.ingestAssistant(head)
	case eventToolCall:
		a.ingestToolCall(head)
	case eventResult:
		a.ingestResult(raw)
	}
}

func (a *streamAccum) finishWithoutResult() (parsedOutput, error) {
	if a.lastDecErr != nil {
		return parsedOutput{}, fmt.Errorf("decode stdout: %w", a.lastDecErr)
	}
	if open := openToolCallCount(a.openToolCalls, a.openAnonymous); open > 0 {
		return parsedOutput{}, fmt.Errorf("stream-json: no terminal result event; %d open tool call(s)", open)
	}
	if a.lastAssistantText != "" {
		return parsedOutput{
			Type:                  eventResult,
			Subtype:               subtypeSuccess,
			Result:                a.lastAssistantText,
			SessionID:             a.lastSessionID,
			ResolvedModel:         a.out.ResolvedModel,
			MissingTerminalResult: true,
		}, nil
	}
	return parsedOutput{}, errors.New("stream-json: no terminal result event")
}

func parseStdout(stdout []byte) (parsedOutput, error) {
	stdout = bytes.TrimSpace(stdout)
	if len(stdout) == 0 {
		return parsedOutput{}, errors.New("empty stdout")
	}
	if single, ok := trySingleResultJSON(stdout); ok {
		return single, nil
	}
	a := &streamAccum{openToolCalls: map[string]struct{}{}}
	for _, raw := range splitNDJSON(stdout) {
		if len(raw) == 0 {
			continue
		}
		a.ingest(raw)
	}
	if !a.gotResult {
		return a.finishWithoutResult()
	}
	return a.out, nil
}

func updateOpenToolCalls(open map[string]struct{}, openAnonymous *int, head streamEventHead) {
	callID := strings.TrimSpace(head.CallID)
	switch head.Subtype {
	case subtypeStarted, subtypeStart:
		if callID == "" {
			*openAnonymous++
			return
		}
		open[callID] = struct{}{}
	case subtypeCompleted, subtypeSuccess, subtypeDone, subtypeFailed, subtypeError:
		if callID == "" {
			if *openAnonymous > 0 {
				*openAnonymous--
			}
			return
		}
		delete(open, callID)
	}
}

func openToolCallCount(open map[string]struct{}, anonymous int) int {
	return len(open) + anonymous
}

func splitNDJSON(b []byte) [][]byte {
	if len(b) == 0 {
		return nil
	}
	out := make([][]byte, 0, 8)
	start := 0
	for i := range b {
		if b[i] != '\n' {
			continue
		}
		line := b[start:i]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			out = append(out, line)
		}
		start = i + 1
	}
	if start < len(b) {
		tail := bytes.TrimSpace(b[start:])
		if len(tail) > 0 {
			out = append(out, tail)
		}
	}
	return out
}
