package convertchoices

import (
	"strings"
	"testing"
)

func TestValidateNamesTheFlagThatIsWrong(t *testing.T) {
	err := Validate(Choices{Level: LevelFull, Test: "mocha"}, Given{Test: true})
	if err == nil || !strings.Contains(err.Error(), "--test") || !strings.Contains(err.Error(), "jest, vitest, keep") {
		t.Fatalf("err = %v", err)
	}
	err = Validate(Choices{Level: LevelConservative, Lint: Smith}, Given{Lint: true})
	if err == nil || !strings.Contains(err.Error(), "--level full") {
		t.Fatalf("a choice needs the full level, got %v", err)
	}
	if err := Validate(Choices{Level: "deep"}, Given{}); err == nil {
		t.Fatal("an unknown level is refused")
	}
	if err := Validate(Choices{Level: LevelFull, Lint: Keep}, Given{Lint: true}); err != nil {
		t.Fatal(err)
	}
}
