package logsearch

import (
	"reflect"
	"testing"
)

func TestFindIgnoresCaseAndColour(t *testing.T) {
	rows := []string{"\x1b[31mFAIL\x1b[0m TestOrders", "ok", "fail fail"}
	got := Find(rows, "fail")
	want := []Match{{Row: 0, From: 0, To: 4}, {Row: 2, From: 0, To: 4}, {Row: 2, From: 5, To: 9}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	if Find(rows, "") != nil {
		t.Fatal("an empty search finds nothing")
	}
	if got := InRow(got, 2); len(got) != 2 {
		t.Fatalf("row 2: %v", got)
	}
}
