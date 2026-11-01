package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestASecretReachesEveryJobThatRunsSmith(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: SecretCredentials([]string{"FETCH_TOKEN"})})
	m := NewModel(c, "dev", "main")
	m.Native = true
	m.AssetPlatforms = []string{"linux-x64"}
	out := RenderGitHub(m)
	const secret = "FETCH_TOKEN: ${{ secrets.FETCH_TOKEN }}"
	if n := strings.Count(out, secret); n < 4 {
		t.Errorf("the secret is handed to setup, verify, native and assets (%d times):\n%s", n, out)
	}
	assets := out[strings.Index(out, "  assets:"):]
	if !strings.Contains(assets, "run: smith ci assets\n        env:\n          SMITH_PLATFORM: ${{ matrix.platform }}\n          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n          "+secret) {
		t.Errorf("the assets build step reads it:\n%s", assets)
	}
	if strings.Count(assets, "GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}") != 1 {
		t.Errorf("GITHUB_TOKEN is not repeated:\n%s", assets)
	}
}

func TestNoSecretsLeavesThePipelineAsItWas(t *testing.T) {
	if got := SecretCredentials(nil); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}
