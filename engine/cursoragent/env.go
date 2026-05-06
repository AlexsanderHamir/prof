package cursoragent

// Adapted from T2A pkgs/agents/runner/cursor/config.go (defaultPassthroughEnvKeys, envPolicy)
// and pkgs/agents/runner/adapterkit/env.go (BuildEnv, IsDeniedEnvKey).

import (
	"os"
	"strings"
)

// envPolicy describes which parent environment keys may reach the cursor-agent child.
type envPolicy struct {
	parentAllowedKeys []string
	extraAllowedKeys  []string
	deniedKeys        []string
	deniedPrefixes    []string
	lookup            func(string) string
}

// defaultPassthroughEnvKeys matches T2A's cursor adapter list for stable cursor-agent execution on Windows and Unix.
var defaultPassthroughEnvKeys = []string{
	"PATH",
	"HOME",
	"USERPROFILE",
	"SYSTEMDRIVE",
	"SYSTEMROOT",
	"WINDIR",
	"COMSPEC",
	"PATHEXT",
	"LOCALAPPDATA",
	"APPDATA",
	"PROGRAMDATA",
	"ALLUSERSPROFILE",
	"PUBLIC",
	"TEMP",
	"TMP",
	"PROGRAMFILES",
	"PROGRAMFILES(X86)",
	"PROGRAMW6432",
	"COMMONPROGRAMFILES",
	"COMMONPROGRAMFILES(X86)",
	"USERNAME",
	"USERDOMAIN",
	"COMPUTERNAME",
	"LOGONSERVER",
	"SESSIONNAME",
	"OS",
	"PROCESSOR_ARCHITECTURE",
	"PROCESSOR_IDENTIFIER",
	"PROCESSOR_LEVEL",
	"PROCESSOR_REVISION",
	"NUMBER_OF_PROCESSORS",
}

func defaultEnvPolicy(extraKeys []string) envPolicy {
	return envPolicy{
		parentAllowedKeys: defaultPassthroughEnvKeys,
		extraAllowedKeys:  append([]string(nil), extraKeys...),
		deniedKeys: []string{
			"DATABASE_URL",
			"OPENAI_API_KEY",
			"ANTHROPIC_API_KEY",
			"GITHUB_TOKEN",
			"AWS_ACCESS_KEY_ID",
			"AWS_SECRET_ACCESS_KEY",
		},
		deniedPrefixes: nil,
		lookup:         nil,
	}
}

func isDeniedEnvKey(key string, policy envPolicy) bool {
	for _, denied := range policy.deniedKeys {
		if key == denied {
			return true
		}
	}
	for _, prefix := range policy.deniedPrefixes {
		if prefix != "" && strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// buildChildEnv assembles an os/exec-style env slice ("KEY=VALUE") from the allowlist only.
func buildChildEnv(reqEnv map[string]string, extraKeys []string) []string {
	policy := defaultEnvPolicy(extraKeys)
	lookup := policy.lookup
	if lookup == nil {
		lookup = os.Getenv
	}
	allowed := make(map[string]struct{}, len(policy.parentAllowedKeys)+len(policy.extraAllowedKeys))
	for _, k := range policy.parentAllowedKeys {
		if k == "" || isDeniedEnvKey(k, policy) {
			continue
		}
		allowed[k] = struct{}{}
	}
	for _, k := range policy.extraAllowedKeys {
		if k == "" || isDeniedEnvKey(k, policy) {
			continue
		}
		allowed[k] = struct{}{}
	}
	merged := map[string]string{}
	for k := range allowed {
		if v := lookup(k); v != "" {
			merged[k] = v
		}
	}
	for k, v := range reqEnv {
		if k == "" || isDeniedEnvKey(k, policy) {
			continue
		}
		merged[k] = v
	}
	out := make([]string, 0, len(merged))
	for k, v := range merged {
		out = append(out, k+"="+v)
	}
	return out
}

func liveHomePaths() []string {
	out := make([]string, 0, 2)
	for _, k := range []string{"HOME", "USERPROFILE"} {
		if v := os.Getenv(k); v != "" {
			out = append(out, v)
		}
	}
	return out
}
