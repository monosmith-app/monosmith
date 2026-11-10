package gowork

import (
	"reflect"
	"testing"
)

func TestParseGoWork(t *testing.T) {
	src := "go 1.24.3\n\nuse (\n\t./apps/api // the api\n\t\"./libs/core\"\n)\nuse ./packages/sdk\n"
	w := ParseGoWork(src)
	if w.GoVersion != "1.24" || !reflect.DeepEqual(w.Use, []string{"apps/api", "libs/core", "packages/sdk"}) {
		t.Fatalf("%+v", w)
	}
}
