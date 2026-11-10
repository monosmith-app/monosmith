package releasing

import (
	"reflect"
	"testing"
)

func TestSameReleaseOptionalDependencies(t *testing.T) {
	manifest := []byte(`{"name":"@acme/cli","optionalDependencies":{"@acme/cli-linux":"1.2.0","@acme/cli-win":"1.2.0","left-pad":"^1.0.0","@acme/old":"1.1.0"}}`)
	if got := SameReleaseOptionalDependencies(manifest, "1.2.0"); !reflect.DeepEqual(got, []string{"@acme/cli-linux", "@acme/cli-win"}) {
		t.Errorf("%v", got)
	}
	if got := SameReleaseOptionalDependencies([]byte(`not json`), "1.2.0"); got != nil {
		t.Errorf("%v", got)
	}
}
