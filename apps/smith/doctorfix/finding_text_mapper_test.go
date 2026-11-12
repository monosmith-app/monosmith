package doctorfix

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFindingTextListsEachFailureWithItsRemedy(t *testing.T) {
	doctor := Finding{
		Step: protocol.FixStep{Description: "check the workspace with doctor", Argv: []string{`C:\bin\smith.exe`, "doctor"}},
		Output: "✓ the repository has commits\r\n" +
			"× api:build writes dist/apps/api but its task caches apps/api/dist\n" +
			"    api:build writes dist/apps/api but its task caches apps/api/dist: a cache hit would leave it empty\n" +
			"    → declare outputs = [\"../../dist/apps/api\"] under [projects.api.tasks.build] in smith.toml\n" +
			"! 1 agent file(s) are missing\n" +
			"    → run `smith agent install`\n" +
			"\n27 passed, 1 failed, 1 warning\n",
	}
	want := "▲ check the workspace with doctor\n" +
		"  × api:build writes dist/apps/api but its task caches apps/api/dist\n" +
		"    → declare outputs = [\"../../dist/apps/api\"] under [projects.api.tasks.build] in smith.toml\n" +
		"  → check again: smith doctor\n"
	if got := FindingText(doctor); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}

	lint := Finding{
		Step: protocol.FixStep{Description: "fix what ESLint can, then lint every project", Argv: []string{"smith", "eslint-fix"}},
		Output: "Project   Fixed    Left  Result\napi         550     871  ▲ 871 errors, 0 warnings\n" +
			"▲ ESLint fixed 10696 problems; 1742 are left in 3 of 3 projects\n→ see them: smith run lint -p api\n",
	}
	want = "▲ fix what ESLint can, then lint every project\n" +
		"  ▲ ESLint fixed 10696 problems; 1742 are left in 3 of 3 projects\n" +
		"    → see them: smith run lint -p api\n" +
		"  → check again: smith eslint-fix\n"
	if got := FindingText(lint); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
