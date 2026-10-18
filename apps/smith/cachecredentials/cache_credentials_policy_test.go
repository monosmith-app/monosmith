package cachecredentials

import (
	"reflect"
	"testing"
)

func names(remote string) []string {
	var out []string
	for _, c := range For(remote) {
		if !c.Tasks || c.Required {
			out = append(out, "!"+c.Env) // every one is optional and given to the tasks
		} else {
			out = append(out, c.Env)
		}
	}
	return out
}

func TestEachBackendNeedsItsOwnCredentials(t *testing.T) {
	cases := map[string][]string{
		"azblob://acme/smith-cache/main": {"SMITH_CACHE_KEY", "AZURE_STORAGE_SAS_TOKEN"},
		"s3://acme/smith":                {"SMITH_CACHE_KEY", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION"},
		"https://cache.acme.dev/smith":   {"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN"},
		"http://localhost:9000/smith":    {"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN"},
	}
	for remote, want := range cases {
		if got := names(remote); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, want %v", remote, got, want)
		}
	}
}

func TestNoRemoteNeedsNothing(t *testing.T) {
	if got := For(""); len(got) != 0 {
		t.Errorf("credentials without a remote: %v", got)
	}
}
