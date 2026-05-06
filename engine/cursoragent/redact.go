package cursoragent

// Adapted from T2A pkgs/agents/runner/adapterkit/redact.go (auth/cookie stripping + home paths).

import (
	"regexp"
	"strings"
)

const redactedValue = "[REDACTED]"

var (
	authHeaderRE   = regexp.MustCompile(`(?i)(authorization:[ \t]*)([^\r\n]+)`)
	cookieHeaderRE = regexp.MustCompile(`(?i)\b((?:set-)?cookie:[ \t]*)([^\r\n]+)`)
)

// Redact applies the same baseline scrubbing as the T2A cursor adapter: Authorization/Cookie
// headers and absolute home paths (replaced with "~"). It does not redact PROF_* assignments.
func Redact(s string, homePaths []string) string {
	if s == "" {
		return s
	}
	out := authHeaderRE.ReplaceAllString(s, "${1}"+redactedValue)
	out = cookieHeaderRE.ReplaceAllString(out, "${1}"+redactedValue)
	for _, hp := range homePaths {
		if hp == "" {
			continue
		}
		out = strings.ReplaceAll(out, hp, "~")
	}
	return out
}
