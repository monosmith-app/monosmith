package xmltext

import "testing"

const project = "<Project>\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <SmithKind>app</SmithKind>\n  </PropertyGroup>\n</Project>\n"

func TestSetElementReplacesTheText(t *testing.T) {
	got, ok := SetElement(project, "PropertyGroup", "SmithKind", "lib")
	if !ok || got != "<Project>\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <SmithKind>lib</SmithKind>\n  </PropertyGroup>\n</Project>\n" {
		t.Fatalf("got %q", got)
	}
}

func TestSetElementAddsAtTheTopOfTheParentWithItsIndent(t *testing.T) {
	got, ok := SetElement(project, "PropertyGroup", "IsPackable", "false")
	want := "<Project>\n  <PropertyGroup>\n    <IsPackable>false</IsPackable>\n    <OutputType>Exe</OutputType>\n    <SmithKind>app</SmithKind>\n  </PropertyGroup>\n</Project>\n"
	if !ok || got != want {
		t.Fatalf("got %q", got)
	}
	if _, ok := SetElement("<a/>", "PropertyGroup", "X", "1"); ok {
		t.Fatal("no parent, nothing added")
	}
}

func TestSetElementBeforePutsItAboveTheNamedElement(t *testing.T) {
	got, ok := SetElementBefore(project, "PropertyGroup", "IsPackable", "false", "SmithKind")
	want := "<Project>\n  <PropertyGroup>\n    <OutputType>Exe</OutputType>\n    <IsPackable>false</IsPackable>\n    <SmithKind>app</SmithKind>\n  </PropertyGroup>\n</Project>\n"
	if !ok || got != want {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveElementTakesItsLine(t *testing.T) {
	got, found := RemoveElement(project, "OutputType")
	if !found || got != "<Project>\n  <PropertyGroup>\n    <SmithKind>app</SmithKind>\n  </PropertyGroup>\n</Project>\n" {
		t.Fatalf("got %q", got)
	}
	if _, found := RemoveElement(project, "Nope"); found {
		t.Fatal("nothing to remove")
	}
}

func TestElementReadsTheText(t *testing.T) {
	if v, ok := Element(project, "SmithKind"); !ok || v != "app" {
		t.Fatalf("got %q %v", v, ok)
	}
}
