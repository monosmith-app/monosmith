package cachemaintenance

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
)

func TestTheMeterCountsWhatARunStoresAndNotWhatItReplays(t *testing.T) {
	store, _ := newStore(t)
	meter := NewWriteMeter(store)
	blob := []byte(strings.Repeat("x", 1000))
	digest := putBlob(t, meter, blob)
	if meter.Bytes() != 1000 || meter.Stored() != 1 {
		t.Fatalf("after a new blob: %d bytes, %d stored", meter.Bytes(), meter.Stored())
	}
	// The same blob again is already there and is not read.
	_ = putBlob(t, meter, blob)
	if meter.Bytes() != 1000 {
		t.Fatalf("a blob that was already stored adds nothing, got %d", meter.Bytes())
	}
	// Pointing a task at its last run is not a result.
	before := meter.Stored()
	if err := taskcache.Remember(meter, "web:build", digest); err != nil {
		t.Fatal(err)
	}
	if meter.Stored() != before || meter.Bytes() != 1000 {
		t.Fatal("last-run pointers are not counted")
	}
	// Reads pass through.
	if ok, _ := meter.Has(taskcache.CAS, digest); !ok {
		t.Fatal("the wrapped store answers reads")
	}
}

func TestTheMeterDoesNotCountAFailedPut(t *testing.T) {
	store, _ := newStore(t)
	meter := NewWriteMeter(store)
	if err := meter.Put(taskcache.CAS, strings.Repeat("a", 64), bytes.NewReader([]byte("not that digest"))); err == nil {
		t.Fatal("a blob that does not match its digest is rejected")
	}
	if meter.Stored() != 0 {
		t.Fatal("a rejected put stored nothing")
	}
}
