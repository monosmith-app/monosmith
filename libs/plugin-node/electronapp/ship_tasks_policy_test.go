package electronapp

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestApplyTasks(t *testing.T) {
	tasks := func() map[string]protocol.Task {
		return map[string]protocol.Task{
			"build":   {Outputs: []string{"../../dist/apps/desk"}, Cache: true},
			"package": {}, "make": {}, "dist": {}, "lint": {},
		}
	}
	t.Run("smith-made", func(t *testing.T) {
		got := tasks()
		ApplyTasks(got, "apps/desk", Setup{Tool: SmithForge}, map[string]string{"package": "node scripts/package.mjs"})
		if strings.Join(got["build"].Outputs, ",") != "../../dist/electron/apps/desk/app" || !got["build"].Cache {
			t.Errorf("build %+v", got["build"])
		}
		if p := got["package"]; strings.Join(p.Ship, ",") != "../../dist/electron/apps/desk/out/*-*-*" || strings.Join(p.Outputs, ",") != "../../dist/apps/desk" {
			t.Errorf("package %+v", p)
		}
	})
	t.Run("forge", func(t *testing.T) {
		got := tasks()
		ApplyTasks(got, "apps/desk", Setup{Tool: Forge, OutDir: "out"}, map[string]string{"package": "electron-forge package", "make": "electron-forge make --arch x64", "lint": "eslint ."})
		if strings.Join(got["package"].Ship, ",") != "out/*-*-*" || strings.Join(got["make"].Ship, ",") != "out/*-*-*,out/make/*" {
			t.Errorf("package %+v make %+v", got["package"], got["make"])
		}
		if len(got["lint"].Ship) != 0 || len(got["build"].Ship) != 0 || strings.Join(got["make"].Outputs, ",") != "../../dist/apps/desk" {
			t.Errorf("only packing tasks ship: %+v", got)
		}
	})
	t.Run("builder", func(t *testing.T) {
		got := tasks()
		ApplyTasks(got, "apps/desk", Setup{Tool: Builder, OutDir: "../../dist/electron/apps/desk"}, map[string]string{"dist": "vite build && electron-builder --win"})
		ship := got["dist"].Ship
		if !slices.Contains(ship, "../../dist/electron/apps/desk/*.exe") || !slices.Contains(ship, "../../dist/electron/apps/desk/*-unpacked") || strings.Join(got["dist"].Outputs, ",") != "../../dist/apps/desk" {
			t.Errorf("dist %+v", got["dist"])
		}
	})
	t.Run("unknown folder ships nothing", func(t *testing.T) {
		got := tasks()
		ApplyTasks(got, "apps/desk", Setup{Tool: Builder}, map[string]string{"dist": "electron-builder"})
		if len(got["dist"].Ship) != 0 {
			t.Errorf("dist %+v", got["dist"])
		}
	})
}
