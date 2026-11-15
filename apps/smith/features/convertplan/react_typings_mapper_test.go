package convertplan

import "testing"

func TestRewriteNxReactTypingsUsesViteClient(t *testing.T) {
	app := `{
  // the app
  "compilerOptions": {
    "types": [
      "node",
      "@nx/react/typings/cssmodule.d.ts",
      "@nx/react/typings/image.d.ts",
      "vite/client"
    ],
    "rootDir": "src"
  }
}`
	got, ok := RewriteNxReactTypings(app)
	want := `{
  // the app
  "compilerOptions": {
    "types": [
      "node",
      "vite/client"
    ],
    "rootDir": "src"
  }
}`
	if !ok || got != want {
		t.Errorf("app: ok %v\n%s", ok, got)
	}
	spec := `{"compilerOptions": {"types": ["jest", "node", "@nx/react/typings/cssmodule.d.ts", "@nx/react/typings/image.d.ts"]}}`
	got, ok = RewriteNxReactTypings(spec)
	if !ok || got != `{"compilerOptions": {"types": ["jest", "node", "vite/client"]}}` {
		t.Errorf("spec: ok %v\n%s", ok, got)
	}
	if _, ok := RewriteNxReactTypings(`{"compilerOptions": {"types": ["node"]}}`); ok {
		t.Error("rewrote a tsconfig without @nx/react typings")
	}
}

func TestPlanNxReactTypingsRewritesViteProjectsAndReportsTheRest(t *testing.T) {
	text := `{"compilerOptions": {"types": ["node", "@nx/react/typings/image.d.ts"]}}`
	plan := &Plan{}
	planNxReactTypings(plan, "apps/ui/tsconfig.app.json", text, []string{"apps/ui/tsconfig.app.json", "apps/ui/vite.config.mts"})
	if plan.Write["apps/ui/tsconfig.app.json"] != `{"compilerOptions": {"types": ["node", "vite/client"]}}` || plan.Count(Changed) != 1 {
		t.Errorf("vite: %+v %+v", plan.Write, plan.Findings)
	}
	plan = &Plan{}
	planNxReactTypings(plan, "apps/web/tsconfig.app.json", text, []string{"apps/web/tsconfig.app.json", "apps/web/webpack.config.js"})
	if len(plan.Write) != 0 || plan.Count(Manual) != 1 {
		t.Errorf("no vite: %+v %+v", plan.Write, plan.Findings)
	}
}
