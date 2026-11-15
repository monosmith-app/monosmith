package doctorview

import (
	"reflect"
	"testing"
)

func TestUnrepeatedKeepsOnlyWhatTheLinesAboveDidNotSay(t *testing.T) {
	said := []string{"golangci-lint is older than CI", `see C:\ws\.golangci.yml for the rules`}
	items := []string{
		"Golangci-lint is older than CI",
		"Install the version CI uses",
		"install  the version   CI uses",
		`  writes C:\ws\.golangci.yml`,
		"  writes ci.yml",
		"  writes ci.yml",
	}
	want := []string{"Install the version CI uses", "  writes ci.yml"}
	if got := Unrepeated(items, said); !reflect.DeepEqual(got, want) {
		t.Errorf("%q, want %q", got, want)
	}
}

func TestTellOnceFoldsMetadataAndShortensARepeatedPath(t *testing.T) {
	detail := "found /opt/bin/lint (lint has version 2.1 built with go1.27 on (unknown)); older than CI"
	how := []string{"smith asks how: keep /opt/bin/lint (lint has version 2.1 built with go1.27 on (unknown)), or update"}
	problems, notes, items := TellOnce("lint matches CI", detail, how)
	if want := []string{"found /opt/bin/lint (lint has version 2.1)", "older than CI"}; !reflect.DeepEqual(problems, want) {
		t.Errorf("problems %q, want %q", problems, want)
	}
	if want := []string{"built with go1.27 on (unknown)"}; !reflect.DeepEqual(notes, want) {
		t.Errorf("notes %q, want %q", notes, want)
	}
	if want := []string{"smith asks how: keep lint (lint has version 2.1), or update"}; !reflect.DeepEqual(items, want) {
		t.Errorf("items %q, want %q", items, want)
	}
}
