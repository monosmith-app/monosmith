package tsconfigchain

import "testing"

func TestRewire(t *testing.T) {
	for name, c := range map[string]struct{ in, want string }{
		"no extends": {
			"{\n  \"compilerOptions\": { \"strict\": true }\n}\n",
			"{\n  \"compilerOptions\": { \"strict\": true },\n  \"extends\": \"../../tsconfig.base.json\"\n}\n",
		},
		"smith preset": {
			"{\n  \"extends\": \"@monosmith/tsconfig/node.json\",\n  \"include\": [\"src\"]\n}\n",
			"{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"include\": [\"src\"]\n}\n",
		},
		"another preset keeps its say": {
			"{\n  \"extends\": \"astro/tsconfigs/strict\"\n}\n",
			"{\n  \"extends\": [\n    \"../../tsconfig.base.json\",\n    \"astro/tsconfigs/strict\"\n  ]\n}\n",
		},
		"comments stay": {
			"{\n  // the project\n  \"extends\": \"../common.json\"\n}\n",
			"{\n  // the project\n  \"extends\": [\n    \"../../tsconfig.base.json\",\n    \"../common.json\"\n  ]\n}\n",
		},
	} {
		got, err := Rewire([]byte(c.in), "apps/web")
		if err != nil || string(got) != c.want {
			t.Errorf("%s:\n%s\nwant:\n%s\n(%v)", name, got, c.want, err)
		}
	}
}
