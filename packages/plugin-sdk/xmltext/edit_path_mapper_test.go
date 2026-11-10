package xmltext

import "testing"

const pom = "<project xmlns=\"http://maven.apache.org/POM/4.0.0\">\n    <artifactId>kit</artifactId>\n    <build>\n        <resources>\n            <resource>\n                <directory>src/main/resources</directory>\n            </resource>\n        </resources>\n    </build>\n</project>\n"

func TestPathReadsOnlyTheDirectChildren(t *testing.T) {
	if _, ok := Path(pom, "project", "build", "directory"); ok {
		t.Fatal("the resource directory is not the build directory")
	}
	if v, ok := Path(pom, "project", "artifactId"); !ok || v != "kit" {
		t.Fatalf("%q %v", v, ok)
	}
}

func TestSetPathAddsTheMissingElementWithoutTouchingTheRest(t *testing.T) {
	got, ok := SetPath(pom, "X", "project", "build", "directory")
	want := "<project xmlns=\"http://maven.apache.org/POM/4.0.0\">\n    <artifactId>kit</artifactId>\n    <build>\n        <resources>\n            <resource>\n                <directory>src/main/resources</directory>\n            </resource>\n        </resources>\n        <directory>X</directory>\n    </build>\n</project>\n"
	if !ok || got != want {
		t.Fatalf("%v\n%s", ok, got)
	}
	if v, _ := Path(got, "project", "build", "directory"); v != "X" {
		t.Fatalf("%q", v)
	}
}

func TestSetPathAddsAWholeBuildBlockAndReplacesAnExistingValue(t *testing.T) {
	src := "<project>\n  <artifactId>kit</artifactId>\n</project>\n"
	got, ok := SetPath(src, "X", "project", "build", "directory")
	want := "<project>\n  <artifactId>kit</artifactId>\n  <build>\n    <directory>X</directory>\n  </build>\n</project>\n"
	if !ok || got != want {
		t.Fatalf("%v\n%s", ok, got)
	}
	again, ok := SetPath(got, "Y", "project", "build", "directory")
	if !ok || again != "<project>\n  <artifactId>kit</artifactId>\n  <build>\n    <directory>Y</directory>\n  </build>\n</project>\n" {
		t.Fatalf("%v\n%s", ok, again)
	}
}

func TestSetPathRefusesTextItCannotEditSafely(t *testing.T) {
	for _, src := range []string{"<project><build/></project>", "<project>", "<other><build></build></other>"} {
		if _, ok := SetPath(src, "X", "project", "build", "directory"); ok {
			t.Errorf("%q must be refused", src)
		}
	}
}
