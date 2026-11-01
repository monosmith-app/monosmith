package configedit

import (
	"strings"
	"testing"
)

func TestFindSaysHowToEditReleaseChannels(t *testing.T) {
	_, err := Find("release.channels")
	if err == nil || !strings.Contains(err.Error(), "[[release.channels]]") {
		t.Fatalf("error = %v", err)
	}
}
