package tracker

import (
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestDetectChangeBetweenTwoObjectsRegression(t *testing.T) {
	base := &parser.LineObj{FnName: "f", Flat: 10, Cum: 10}
	cur := &parser.LineObj{FnName: "f", Flat: 20, Cum: 20}
	r, err := detectChangeBetweenTwoObjects(base, cur)
	if err != nil {
		t.Fatal(err)
	}
	if r.ChangeType != internal.REGRESSION {
		t.Fatalf("got %s", r.ChangeType)
	}
}

func TestDetectChangeBetweenTwoObjectsImprovement(t *testing.T) {
	base := &parser.LineObj{FnName: "f", Flat: 20, Cum: 20}
	cur := &parser.LineObj{FnName: "f", Flat: 10, Cum: 10}
	r, err := detectChangeBetweenTwoObjects(base, cur)
	if err != nil {
		t.Fatal(err)
	}
	if r.ChangeType != internal.IMPROVEMENT {
		t.Fatalf("got %s", r.ChangeType)
	}
}

func TestDetectChangeNilBaseline(t *testing.T) {
	cur := &parser.LineObj{FnName: "f", Flat: 1, Cum: 1}
	_, err := detectChangeBetweenTwoObjects(nil, cur)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLineObjByShortName(t *testing.T) {
	objs := []*parser.LineObj{{FnName: "a", Flat: 1}, {FnName: "b", Flat: 2}}
	m := lineObjByShortName(objs)
	if len(m) != 2 || m["a"].Flat != 1 {
		t.Fatal(m)
	}
}
