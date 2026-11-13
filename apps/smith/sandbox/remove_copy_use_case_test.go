package sandbox

import (
	"errors"
	"testing"
	"time"
)

// A folder a lingering process still holds is deleted once that process lets
// go, and a folder that stays busy is given up on after a bounded wait.
func TestRemoveCopyRetriesWhileTheFolderIsBusy(t *testing.T) {
	busy := errors.New("the process cannot access the file because it is being used by another process")
	calls, slept := 0, time.Duration(0)
	sleep := func(d time.Duration) { slept += d }
	err := removeCopy("x", func(string) error {
		calls++
		if calls < 3 {
			return busy
		}
		return nil
	}, sleep)
	if err != nil || calls != 3 {
		t.Fatalf("removed after %d calls: %v", calls, err)
	}
	calls, slept = 0, 0
	err = removeCopy("x", func(string) error { calls++; return busy }, sleep)
	if !errors.Is(err, busy) || slept > 6*time.Second || calls < 2 {
		t.Fatalf("gave up after %d calls and %s: %v", calls, slept, err)
	}
}
