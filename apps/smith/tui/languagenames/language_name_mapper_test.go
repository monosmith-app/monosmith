package languagenames

import "testing"

func TestIdsReadAsPeopleSayThem(t *testing.T) {
	if Of("jvm") != "Java / Kotlin" || Of("dotnet") != "C# / .NET" || Of("klingon") != "klingon" {
		t.Fatal("names")
	}
	got := Order([]string{"bicep", "dotnet", "go", "node"})
	if got[0] != "node" || got[1] != "go" || got[2] != "dotnet" || got[3] != "bicep" {
		t.Fatalf("popular first: %v", got)
	}
}
