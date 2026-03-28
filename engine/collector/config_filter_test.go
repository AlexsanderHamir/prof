package collector

import (
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

func TestGlobalFilterFromConfig(t *testing.T) {
	cfg := &internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			internal.GlobalSign: {IncludePrefixes: []string{"p"}},
		},
	}
	f, ok := globalFilterFromConfig(cfg)
	if !ok || len(f.IncludePrefixes) != 1 {
		t.Fatalf("got %#v %v", f, ok)
	}
	cfg2 := &internal.Config{}
	_, ok = globalFilterFromConfig(cfg2)
	if ok {
		t.Fatal("expected no global")
	}
}

func TestResolveFunctionFilter(t *testing.T) {
	global := internal.FunctionFilter{IncludePrefixes: []string{"g"}}
	cfg := &internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			internal.GlobalSign: global,
		},
	}
	if f := resolveFunctionFilter(cfg, "any", global); len(f.IncludePrefixes) != 1 {
		t.Fatal(f)
	}
	cfgLocal := &internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			"cpu": {IgnoreFunctions: []string{"init"}},
		},
	}
	f2 := resolveFunctionFilter(cfgLocal, "cpu", global)
	if len(f2.IgnoreFunctions) != 1 {
		t.Fatal(f2)
	}
}
