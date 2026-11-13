package discovery

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGradleFixesWhatKtlintAndAndroidLintCanFix(t *testing.T) {
	for _, tc := range []struct {
		lint []string
		want []string
	}{
		{[]string{"ktlintCheck"}, []string{"ktlintFormat"}},
		{[]string{"lint"}, []string{"lintFix"}},
		{[]string{"checkstyleMain", "checkstyleTest"}, nil},
	} {
		if got := gradleFixTasks(tc.lint); !slices.Equal(got, tc.want) {
			t.Errorf("%v: got %v, want %v", tc.lint, got, tc.want)
		}
	}
	tasks := GradleTasks("apps/api", GradleBuild{Lint: []string{"checkstyleMain"}}, protocol.ProjectApp)
	if tasks["lint"].Fix != nil {
		t.Fatalf("Checkstyle has no fix: %+v", tasks["lint"].Fix)
	}
}
