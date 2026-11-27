package tsconfigchain

import "testing"

func TestAliasedMatchesWildcardAndExactKeysOnly(t *testing.T) {
	aliases := []string{"@shared/*", "@integrations", "~/*"}
	for name, want := range map[string]bool{
		"@shared/contracts": true, "@integrations": true, "~": true,
		"@integrations-x": false, "@sharedx/a": false, "react": false,
	} {
		if got := Aliased(name, aliases); got != want {
			t.Errorf("Aliased(%q) = %v, want %v", name, got, want)
		}
	}
}
