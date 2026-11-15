package convertplan

import (
	"reflect"
	"testing"
)

func TestAzurePipelineSettingsReadsThePoolAndTheVariableGroups(t *testing.T) {
	cases := map[string]struct {
		text   string
		pool   string
		groups []string
	}{
		"a named pool and a group": {
			text: "trigger:\n  branches:\n    include:\n      - main\n\npool:\n  name: AzurePipelineManagedPool-Windows\n\nvariables:\n  # holds the npm secret\n  - group: Build\n  - name: npm_config_cache\n    value: $(Pipeline.Workspace)/.npm\n\nsteps:\n  - script: npm ci\n",
			pool: "AzurePipelineManagedPool-Windows", groups: []string{"Build"},
		},
		"a hosted image is smith's default": {
			text: "pool:\n  vmImage: ubuntu-latest\n\nsteps:\n  - script: echo\n",
		},
		"several groups, quoted and with a comment": {
			text:   "variables:\n  - group: 'Build'\n  - group: \"Release secrets\"   # for publishing\n  - group: Build\n",
			groups: []string{"Build", "Release secrets"},
		},
		"a name in a later block is not the pool's": {
			text: "pool:\n  vmImage: ubuntu-latest\n\nsteps:\n  - task: Cache@2\n    inputs:\n      name: not-a-pool\n",
		},
		"CRLF": {
			text: "pool:\r\n  name: Managed\r\n\r\nvariables:\r\n  - group: Build\r\n",
			pool: "Managed", groups: []string{"Build"},
		},
		"nothing": {text: "steps:\n  - script: echo\n"},
	}
	for name, c := range cases {
		pool, groups := AzurePipelineSettings(c.text)
		if pool != c.pool || !reflect.DeepEqual(groups, c.groups) {
			t.Errorf("%s: pool %q groups %q, want %q %q", name, pool, groups, c.pool, c.groups)
		}
	}
}
