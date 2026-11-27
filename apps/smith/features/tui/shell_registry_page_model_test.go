package tui

import (
	"bytes"
	"encoding/base64"
	"testing"
)

const axiosPage = "https://www.npmjs.com/package/axios"

func TestGOpensTheRegistryPageOfTheSelectedDependencyAndCopiesTheLink(t *testing.T) {
	l := newLab(t, true, 140, 34)
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh
	l.keys("5")
	l.has("Registry: "+axiosPage, "g opens it in your browser")
	if len(l.opened) != 0 {
		t.Fatalf("a browser opened before any key: %v", l.opened)
	}
	l.keys("g")
	if len(l.opened) != 1 || l.opened[0] != axiosPage {
		t.Fatalf("opened %v", l.opened)
	}
	if got := out.String(); got != "\x1b]52;c;"+base64.StdEncoding.EncodeToString([]byte(axiosPage))+"\a" {
		t.Fatalf("the link is copied too: %q", got)
	}
}

func TestGOnAPackageWithoutAPageOpensNothingAndSaysWhy(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("5", "down", "g") // zod has no page in the sample
	if len(l.opened) != 0 {
		t.Fatalf("opened %v", l.opened)
	}
	l.has("No registry page is known for zod")
}

func TestShiftYCopiesOnlyTheLinkOnTheDepsView(t *testing.T) {
	l := newLab(t, true, 140, 34)
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh
	l.keys("5", "Y")
	if len(l.opened) != 0 {
		t.Fatalf("copying must not open a browser: %v", l.opened)
	}
	if got := out.String(); got != "\x1b]52;c;"+base64.StdEncoding.EncodeToString([]byte(axiosPage))+"\a" {
		t.Fatalf("copied %q", got)
	}
}
