package convertchoices

import "testing"

func TestFlagsSpellTheChoices(t *testing.T) {
	got := Flags(Choices{Level: LevelFull, Lint: Smith, Test: Jest, Build: Smith, CI: Keep})
	if got != "--level full --lint smith --test jest --build smith --ci keep" {
		t.Fatal(got)
	}
}
