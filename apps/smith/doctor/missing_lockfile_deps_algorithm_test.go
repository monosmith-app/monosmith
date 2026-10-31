package doctor

import (
	"reflect"
	"testing"
)

func TestMissingLockfileDepsFindsAnOptionalPackageTheLockfileDoesNotHold(t *testing.T) {
	lock := `{"packages":{"":{},
	"node_modules/@napi-rs/wasm-runtime":{"dependencies":{"@emnapi/core":"1","@emnapi/runtime":"1"}},
	"node_modules/@emnapi/core":{},
	"node_modules/esbuild":{"optionalDependencies":{"@esbuild/win32-x64":"1","@esbuild/linux-x64":"1"}},
	"node_modules/@esbuild/win32-x64":{}}}`
	got, err := missingLockfileDeps([]byte(lock))
	if err != nil || !reflect.DeepEqual(got, []string{"@emnapi/runtime", "@esbuild/linux-x64"}) {
		t.Fatalf("%v %v", got, err)
	}
}

func TestMissingLockfileDepsResolvesNestedAndWorkspaceAndBundledPackages(t *testing.T) {
	lock := `{"packages":{"":{},
	"packages/a":{"dependencies":{"x":"1"}},
	"node_modules/x":{},
	"node_modules/p":{"dependencies":{"y":"1","z":"1"},"bundleDependencies":["z"]},
	"node_modules/p/node_modules/y":{"dependencies":{"x":"1"}}}}`
	got, err := missingLockfileDeps([]byte(lock))
	if err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
}

func TestMissingLockfileDepsRejectsInvalidJSON(t *testing.T) {
	if _, err := missingLockfileDeps([]byte("{")); err == nil {
		t.Fatal("want an error")
	}
}
