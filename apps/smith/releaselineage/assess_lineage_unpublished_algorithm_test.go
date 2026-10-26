package releaselineage

import (
	"strings"
	"testing"
)

func TestARegistryThatAnswers401Or404ForANewPackageIsNotANetworkFault(t *testing.T) {
	for _, detail := range []string{
		"npm error code E401",
		"npm error code E404",
		"npm error 401 Unauthorized - GET https://pkgs.dev.azure.com/org/_packaging/feed/npm/registry/demo-pkg",
		"404 Not Found",
	} {
		got := Assess([]Subject{{Project: "demo", Prefix: "demo@", Unreachable: true, Registry: "https://pkgs.dev.azure.com/x", Detail: detail}})
		if len(got) != 0 {
			t.Errorf("%q is a package that is not published yet, got %+v", detail, got)
		}
	}
	for _, detail := range []string{"connection refused", "npm error code ECONNRESET", "port 4040 refused", "timeout after 15000ms"} {
		got := Assess([]Subject{{Project: "demo", Prefix: "demo@", Unreachable: true, Detail: detail}})
		if len(got) != 1 || !strings.Contains(got[0].Message, "could not ask") {
			t.Errorf("%q is a network fault, got %+v", detail, got)
		}
	}
}
