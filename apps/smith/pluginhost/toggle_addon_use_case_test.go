package pluginhost

import (
	"strings"
	"testing"
)

func TestToggleExplainsAnAddonCannotBeDisabled(t *testing.T) {
	_, err := Toggle(nil, []string{"eslint"}, false)
	if err == nil || !strings.Contains(err.Error(), "add-on plugin") {
		t.Fatalf("error = %v", err)
	}
	if _, err := Toggle(nil, []string{"nope"}, false); err == nil || strings.Contains(err.Error(), "add-on") {
		t.Fatalf("an unknown name is not an add-on: %v", err)
	}
}
