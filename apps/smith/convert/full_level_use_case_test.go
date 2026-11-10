package convert

import (
	"errors"
	"testing"
)

func TestCommandForNamesTheShimNotThePlatformBinary(t *testing.T) {
	notFound := func(string) (string, error) { return "", errors.New("not found") }
	platform := "C:\\Users\\me\\AppData\\Roaming\\npm\\node_modules\\monosmith\\node_modules\\@monosmith\\smith-win32-x64\\bin\\smith.exe"
	if got := commandFor(platform, notFound); got != "smith" {
		t.Errorf("an npm platform binary is named %q", got)
	}
	if got := commandFor("/home/me/build/smith", notFound); got != "/home/me/build/smith" {
		t.Errorf("a binary that is not on the path keeps its path, got %q", got)
	}
}
