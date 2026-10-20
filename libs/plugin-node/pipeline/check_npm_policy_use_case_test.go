package pipeline

import (
	"reflect"
	"testing"
)

const policyLock = `{"lockfileVersion":3,"packages":{
 "":{"name":"root"},
 "apps/web":{"name":"web"},
 "node_modules/web":{"resolved":"apps/web","link":true},
 "node_modules/good":{"resolved":"https://registry.npmjs.org/good/-/good-1.0.0.tgz","integrity":"sha512-x"},
 "node_modules/nohash":{"resolved":"https://registry.npmjs.org/nohash/-/nohash-1.0.0.tgz"},
 "node_modules/viagit":{"resolved":"git+ssh://git@github.com/a/viagit.git#abc"},
 "node_modules/plain":{"resolved":"http://example.com/plain.tgz","integrity":"sha512-y"},
 "node_modules/esbuild":{"resolved":"https://registry.npmjs.org/esbuild/-/esbuild-1.tgz","integrity":"sha512-z","hasInstallScript":true},
 "node_modules/a/node_modules/sneaky":{"resolved":"https://registry.npmjs.org/sneaky/-/s.tgz","integrity":"sha512-q","hasInstallScript":true}
}}`

func TestCheckNpmPolicyReportsEachViolationSorted(t *testing.T) {
	got, err := CheckNpmPolicy([]byte(policyLock), []string{"esbuild"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"nohash has no integrity hash in the lockfile",
		"plain is fetched over plain http (http://example.com/plain.tgz)",
		"sneaky runs an install script and is not in audit_allow_install_scripts",
		"viagit is fetched from git (git+ssh://git@github.com/a/viagit.git#abc): depend on a published version",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%q", got)
	}
}

func TestCheckNpmPolicyAcceptsACleanLockfile(t *testing.T) {
	got, err := CheckNpmPolicy([]byte(`{"packages":{"":{},"node_modules/ok":{"resolved":"https://r/ok.tgz","integrity":"sha512-x"}}}`), nil)
	if err != nil || len(got) != 0 {
		t.Errorf("%v %v", got, err)
	}
}

func TestCheckNpmPolicyRejectsALockfileItCannotRead(t *testing.T) {
	if _, err := CheckNpmPolicy([]byte(`{"lockfileVersion":1}`), nil); err == nil {
		t.Error("lockfile v1 has no packages table")
	}
	if _, err := CheckNpmPolicy([]byte(`nope`), nil); err == nil {
		t.Error("not JSON")
	}
}
