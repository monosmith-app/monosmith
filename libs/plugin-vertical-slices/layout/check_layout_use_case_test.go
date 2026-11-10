package layout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestCheckGo(t *testing.T) {
	root := tree(t,
		"go.mod", "main.go", "doc.go", "main_windows.go",
		"greeting/doc.go", "greeting/greet_use_case.go", "greeting/greet_use_case_test.go", "greeting/run_process_use_case_windows.go",
		"greeting/stuff.go", "server.go",
		"utils/x_mapper.go",
		"testdata/whatever.go", ".hidden/bad.go", "vendor/bad.go",
	)
	r, err := Check(root, "go")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{"greeting/stuff.go - a Go file is", "server.go - only doc.go and main.go", `utils/ - "utils" is a junk drawer`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 3 {
		t.Errorf("want 3 problems, got:\n%s", got)
	}
}

func TestCheckPython(t *testing.T) {
	root := tree(t,
		"pyproject.toml", "setup.py",
		"src/kit/__init__.py", "src/kit/__main__.py", "src/kit/cli.py",
		"src/kit/greeting/__init__.py", "src/kit/greeting/greet_use_case.py", "src/kit/greeting/test_greet_use_case.py", "src/kit/greeting/stuff.py",
		"src/kit/greeting/__main__.py",
		"src/kit/BadName/__init__.py", "src/kit/helpers/__init__.py",
		"tests/conftest.py", "tests/test_greet_use_case.py", "tests/fixtures/sample.py",
		".venv/lib/x.py", "src/kit.egg-info/x.py",
	)
	r, err := Check(root, "python")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/kit/cli.py - only __init__.py and __main__.py",
		"src/kit/greeting/stuff.py - a Python module is",
		"src/kit/greeting/__main__.py - a Python module is",
		"src/kit/BadName/ - a Python package is snake_case",
		`src/kit/helpers/ - "helpers" is a junk drawer`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckDart(t *testing.T) {
	root := tree(t,
		"pubspec.yaml",
		"lib/kit.dart", "lib/main.dart",
		"lib/src/greeting/greeting.dart", "lib/src/greeting/greet_use_case.dart", "lib/src/greeting/greeting_model.g.dart",
		"lib/src/greeting/stuff.dart", "lib/src/loose.dart", "lib/src/Orders/place_order_use_case.dart", "lib/src/helpers/x_mapper.dart",
		"bin/main.dart",
		"test/src/greeting/greet_use_case_test.dart", "test/widget_test.dart", "test/testing/fake_clock.dart", "test/sample.dart",
		"integration_test/app_test.dart",
		"android/app/src/main/kotlin/Main.dart", "example/lib/whatever.dart", ".dart_tool/x.dart", "build/x.dart",
	)
	r, err := Check(root, "dart")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"lib/src/greeting/stuff.dart - a Dart file in a slice is <name>_<role>.dart",
		"lib/src/loose.dart - lib/src holds slice folders",
		"lib/src/Orders/ - a Dart folder is lowercase_with_underscores",
		`lib/src/helpers/ - "helpers" is a junk drawer`,
		"test/sample.dart - a test is <basename>_test.dart",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckRust(t *testing.T) {
	root := tree(t,
		"Cargo.toml", "build.rs",
		"src/main.rs", "src/lib.rs", "src/loose.rs",
		"src/greeting/mod.rs", "src/greeting/greet_use_case.rs", "src/greeting/stuff.rs",
		"src/orders.rs", "src/orders/place_order_use_case.rs", "src/BadName/mod.rs",
		"src/bin/admin_tool.rs",
		"tests/greet_test.rs", "tests/testing/fake_clock.rs", "tests/smoke.rs",
		"benches/x.rs", "target/debug/x.rs",
	)
	r, err := Check(root, "rust")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/loose.rs - src/ holds main.rs, lib.rs and slice modules",
		"src/greeting/stuff.rs - a Rust file in a slice is <name>_<role>.rs",
		"src/BadName/ - a Rust module folder is snake_case",
		"tests/smoke.rs - an integration test is <name>_test.rs",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 4 {
		t.Errorf("want 4 problems, got:\n%s", got)
	}
}

func TestCheckRuby(t *testing.T) {
	root := tree(t,
		"kit.gemspec", "Gemfile", "exe/kit", "Rakefile.rb",
		"lib/kit.rb", "lib/kit/version.rb", "lib/kit/loose.rb",
		"lib/kit/greeting/greeting.rb", "lib/kit/greeting/greet_use_case.rb", "lib/kit/greeting/stuff.rb",
		"lib/kit/orders.rb", "lib/kit/orders/place_order_use_case.rb", "lib/kit/BadName/x_mapper.rb",
		"spec/spec_helper.rb", "spec/kit/greeting/greet_use_case_spec.rb", "spec/support/fake_clock.rb", "spec/sample.rb",
		"test/test_helper.rb", "test/kit/greet_use_case_test.rb",
	)
	r, err := Check(root, "ruby")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"lib/kit/loose.rb - the gem's namespace folder holds version.rb",
		"lib/kit/greeting/stuff.rb - a Ruby file in a slice is <name>_<role>.rb",
		"lib/kit/BadName/ - a Ruby folder is snake_case",
		"spec/sample.rb - a spec is <basename>_spec.rb",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 4 {
		t.Errorf("want 4 problems, got:\n%s", got)
	}
}

func TestCheckCpp(t *testing.T) {
	root := tree(t,
		"CMakeLists.txt", "cmake/x.cpp", "third_party/lib/util.cpp",
		"src/main.cpp", "src/loose.cpp",
		"src/greeting/greet_use_case.cpp", "src/greeting/stuff.cpp",
		"src/orders/PlaceOrderUseCase.cpp", "src/orders/PlaceOrderUseCase.h", "src/orders/OrderTotalPolicy.hpp",
		"src/mixed/place_order_use_case.cpp", "src/mixed/OrderTotalPolicy.cpp",
		"include/kit/greeting.hpp", "include/kit/greeting/greet_use_case.hpp", "include/kit/extra.hpp",
		"tests/greeting/greet_use_case_test.cpp", "tests/orders/PlaceOrderUseCaseTest.cpp", "tests/testing/fake_clock.hpp", "tests/sample.cpp",
	)
	r, err := Check(root, "cpp")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/loose.cpp - src/ holds main.cpp and slice folders",
		"src/greeting/stuff.cpp - a C++ file in a slice is <name>_<role> or <Name><Role>",
		"src/mixed/OrderTotalPolicy.cpp - this folder mixes snake_case",
		"include/kit/extra.hpp - include/<target>/ holds the public headers",
		"tests/sample.cpp - a test is <name>_test.cpp or <Name>Test.cpp",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckDotnet(t *testing.T) {
	root := tree(t,
		"Kit.csproj", "Program.cs", "GlobalUsings.cs", "Loose.cs",
		"Greeting/Greeting.cs", "Greeting/GreetUseCase.cs", "Greeting/Stuff.cs", "Greeting/GreetForm.Designer.cs",
		"Orders/IOrderRepository.cs", "orders_lower/PlaceOrderUseCase.cs", "Helpers/FormatMapper.cs",
		"tests/Kit.Tests.csproj", "tests/Greeting/GreetUseCaseTests.cs", "tests/Testing/FakeClock.cs", "tests/Sample.cs",
		"bin/Debug/x.cs", "obj/x.cs", "Properties/AssemblyInfo.cs",
	)
	r, err := Check(root, "dotnet")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"Loose.cs - the project's root holds Program.cs",
		"Greeting/Stuff.cs - a file in a slice is <Name><Role>.cs",
		"orders_lower/ - a .NET folder is a PascalCase namespace",
		`Helpers/ - "Helpers" is a junk drawer`,
		"tests/Sample.cs - a test is <Name>Tests.cs",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckJvm(t *testing.T) {
	root := tree(t,
		"build.gradle.kts", "settings.gradle.kts", "buildSrc/src/main/kotlin/Conventions.kt",
		"src/main/kotlin/com/acme/kit/Main.kt", "src/main/kotlin/com/acme/kit/MainActivity.kt",
		"src/main/kotlin/com/acme/kit/greeting/GreetUseCase.kt", "src/main/kotlin/com/acme/kit/greeting/Stuff.kt",
		"src/main/java/com/acme/kit/orders/OrderTotalPolicy.java", "src/main/java/com/acme/kit/BadPkg/PlaceOrderUseCase.java",
		"src/main/kotlin/com/acme/kit/util/FormatMapper.kt", "src/main/res/layout/x.kt",
		"src/test/kotlin/com/acme/kit/greeting/GreetUseCaseTest.kt", "src/test/kotlin/com/acme/kit/testing/FakeClock.kt",
		"src/test/kotlin/com/acme/kit/Sample.kt", "src/androidTest/kotlin/com/acme/kit/GreetIT.kt",
	)
	r, err := Check(root, "jvm")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/main/kotlin/com/acme/kit/greeting/Stuff.kt - a class in a slice is <Name><Role>",
		"src/main/java/com/acme/kit/BadPkg/ - a package is lowercase",
		`src/main/kotlin/com/acme/kit/util/ - "util" is a junk drawer`,
		"src/test/kotlin/com/acme/kit/Sample.kt - a test is <Name>Test",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 4 {
		t.Errorf("want 4 problems, got:\n%s", got)
	}
}

func TestCheckPhp(t *testing.T) {
	root := tree(t,
		"composer.json", "index.php", "config/app.php", "public/index.php",
		"src/Kernel.php", "src/Loose.php",
		"src/Greeting/GreetUseCase.php", "src/Greeting/Stuff.php", "src/orders/PlaceOrderUseCase.php", "src/Helpers/FormatMapper.php",
		"tests/bootstrap.php", "tests/Greeting/GreetUseCaseTest.php", "tests/Testing/FakeClock.php", "tests/Sample.php",
	)
	r, err := Check(root, "php")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/Loose.php - src/ holds the slice folders",
		"src/Greeting/Stuff.php - a class in a slice is <Name><Role>.php",
		"src/orders/ - a PHP folder is a PascalCase namespace",
		`src/Helpers/ - "Helpers" is a junk drawer`,
		"tests/Sample.php - a test is <Name>Test.php",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckWarnsAboutLargeSlices(t *testing.T) {
	var files []string
	for i := range SliceSize + 1 {
		files = append(files, fmt.Sprintf("big/step%c_use_case.go", 'a'+i), fmt.Sprintf("big/step%c_use_case_test.go", 'a'+i))
	}
	r, err := Check(tree(t, append(files, "big/doc.go", "small/one_model.go")...), "go")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK() || len(r.Warnings) != 1 || !strings.HasPrefix(r.Warnings[0], "big/ - 13 production files") {
		t.Fatalf("%+v", r)
	}
}

func TestCheckRejectsOtherLanguages(t *testing.T) {
	if _, err := Check(t.TempDir(), "node"); err == nil {
		t.Fatal("node is checked by @monosmith/eslint-config, not here")
	}
}

func TestCheckRejectsABannedWordInAFileName(t *testing.T) {
	for _, c := range []struct {
		language string
		files    []string
		flagged  []string
	}{
		{"go", []string{"go.mod", "main.go", "greeting/doc.go", "greeting/parse_helper_mapper.go", "greeting/data_mapper_test.go", "greeting/greet_use_case.go", "greeting/metadata_mapper.go", "testdata/util_mapper.go", "fixtures/helper_mapper.go"},
			[]string{`greeting/parse_helper_mapper.go - "helper" is in the name`, `greeting/data_mapper_test.go - "data" is in the name`}},
		{"python", []string{"pyproject.toml", "src/kit/__init__.py", "src/kit/greeting/__init__.py", "src/kit/greeting/shared_state_store.py", "src/kit/greeting/greet_use_case.py"},
			[]string{`shared_state_store.py - "shared" is in the name`}},
		{"ruby", []string{"lib/kit.rb", "lib/kit/orders/orders.rb", "lib/kit/orders/misc_policy.rb", "spec/spec_helper.rb", "spec/orders/place_order_use_case_spec.rb"},
			[]string{`lib/kit/orders/misc_policy.rb - "misc" is in the name`}},
	} {
		t.Run(c.language, func(t *testing.T) {
			r, err := Check(tree(t, c.files...), c.language)
			if err != nil {
				t.Fatal(err)
			}
			got := strings.Join(r.Problems, "\n")
			for _, want := range c.flagged {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q in:\n%s", want, got)
				}
			}
			if n := strings.Count(got, "names no outcome"); n != len(c.flagged) {
				t.Errorf("want %d file-name problems, got %d:\n%s", len(c.flagged), n, got)
			}
		})
	}
}
