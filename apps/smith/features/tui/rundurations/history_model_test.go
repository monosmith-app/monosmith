package rundurations

import (
	"reflect"
	"testing"
	"time"
)

func TestATaskKeepsItsLastFiveDurations(t *testing.T) {
	h := History{}
	for i := 1; i <= 7; i++ {
		h = h.Record("web:build", time.Duration(i)*time.Second)
	}
	if got := h["web:build"]; !reflect.DeepEqual(got, []int64{3000, 4000, 5000, 6000, 7000}) {
		t.Fatalf("%v", got)
	}
	if got := h.Record("web:build", 0); !reflect.DeepEqual(got, h) {
		t.Fatal("no time is no record")
	}
	before := History{"a": {1}}
	_ = before.Record("a", time.Second)
	if len(before["a"]) != 1 {
		t.Fatal("Record changes no history it was given")
	}
}

func TestTheTypicalTimeIsTheMedian(t *testing.T) {
	h := History{"odd": {1000, 9000, 2000}, "even": {1000, 3000, 2000, 9000}}
	if d, ok := h.Typical("odd"); !ok || d != 2*time.Second {
		t.Fatalf("odd: %v %v", d, ok)
	}
	if d, ok := h.Typical("even"); !ok || d != 2500*time.Millisecond {
		t.Fatalf("even: %v %v", d, ok)
	}
	if _, ok := h.Typical("none"); ok {
		t.Fatal("a task without history has no typical time")
	}
}
