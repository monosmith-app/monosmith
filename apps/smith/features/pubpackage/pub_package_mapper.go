package pubpackage

import (
	"fmt"
	"sort"
	"strings"
)

// dartOS and dartArch spell Go's os and arch the way Dart's Abi names do
// (Abi.macosArm64), which the launcher matches.
var (
	dartOS   = map[string]string{"linux": "linux", "darwin": "macos", "windows": "windows"}
	dartArch = map[string]string{"amd64": "X64", "arm64": "Arm64"}
)

// Platform is where an archive's binary goes in the package (macos-arm64)
// and the Dart Abi it runs on (macosArm64); ok is false for a platform Dart
// does not run on.
func Platform(goos, goarch string) (dir, abi string, ok bool) {
	os, okOS := dartOS[goos]
	arch, okArch := dartArch[goarch]
	if !okOS || !okArch {
		return "", "", false
	}
	return os + "-" + strings.ToLower(arch), os + arch, true
}

// Pubspec is the package's pubspec.yaml: its one executable runs the
// launcher in bin/.
func Pubspec(r Release) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", r.Name)
	fmt.Fprintf(&b, "description: %s\n", quote(r.Description))
	fmt.Fprintf(&b, "version: %s\n", r.Version)
	if r.Homepage != "" {
		fmt.Fprintf(&b, "homepage: %s\n", r.Homepage)
	}
	if r.Repository != "" {
		fmt.Fprintf(&b, "repository: %s\n", r.Repository)
	}
	fmt.Fprintf(&b, "\nenvironment:\n  sdk: ^3.0.0\n\nexecutables:\n  %s: %s\n", r.Binary, r.Binary)
	return b.String()
}

// quote writes a YAML double-quoted string.
func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// Launcher is bin/<binary>.dart: it finds the package's own files, picks the
// binary built for the running platform and runs it with the same
// arguments, standard streams and exit code. abis maps each Dart Abi name
// to the binary's directory under lib/src/bin.
func Launcher(r Release, abis map[string]string) string {
	names := make([]string, 0, len(abis))
	for abi := range abis {
		names = append(names, abi)
	}
	sort.Strings(names)
	var entries strings.Builder
	for _, abi := range names {
		fmt.Fprintf(&entries, "  Abi.%s: '%s',\n", abi, abis[abi])
	}
	return strings.NewReplacer("__BINARY__", r.Binary, "__PACKAGE__", r.Name, "__PLATFORMS__", entries.String()).Replace(`// Runs the __BINARY__ binary this package ships for the running platform;
// nothing is downloaded after the package is activated.
import 'dart:ffi' show Abi;
import 'dart:io';
import 'dart:isolate';

const platforms = <Abi, String>{
__PLATFORMS__};

Future<void> main(List<String> args) async {
  final platform = platforms[Abi.current()];
  if (platform == null) {
    stderr.writeln('__BINARY__ has no build for ${Abi.current()}');
    exit(1);
  }
  final library = await Isolate.resolvePackageUri(Uri.parse('package:__PACKAGE__/__PACKAGE__.dart'));
  if (library == null) {
    stderr.writeln('__BINARY__: the __PACKAGE__ package cannot be found');
    exit(1);
  }
  final name = Platform.isWindows ? '__BINARY__.exe' : '__BINARY__';
  final binary = File.fromUri(library.resolve('src/bin/$platform/$name'));
  if (!Platform.isWindows) {
    // The pub cache may not keep the executable bit.
    await Process.run('chmod', ['+x', binary.path]);
  }
  final process = await Process.start(binary.path, args, mode: ProcessStartMode.inheritStdio);
  exit(await process.exitCode);
}
`)
}

// Library is lib/<package>.dart: the launcher finds the package's files
// through it.
func Library(r Release) string {
	return fmt.Sprintf("/// %s ships the %s command: `dart pub global activate %s`, then run\n/// `%s`. The package has no Dart API.\nlibrary;\n", r.Name, r.Binary, r.Name, r.Binary)
}

// Readme is the package's README.md, shown on pub.dev.
func Readme(r Release) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", r.Name)
	if r.Description != "" {
		b.WriteString(r.Description + "\n\n")
	}
	fmt.Fprintf(&b, "```sh\ndart pub global activate %s   # or: flutter pub global activate %s\n%s --version\n```\n\n", r.Name, r.Name, r.Binary)
	fmt.Fprintf(&b, "The package ships the prebuilt %s binary for every platform it supports and runs the one for yours; nothing is downloaded after activation. It has no Dart API.\n", r.Binary)
	if r.Homepage != "" {
		fmt.Fprintf(&b, "\nMore: %s\n", r.Homepage)
	}
	return b.String()
}

// Changelog is the package's CHANGELOG.md, which pub.dev asks for.
func Changelog(r Release) string {
	link := ""
	if r.Repository != "" {
		link = fmt.Sprintf(" The release notes are on %s/releases.", r.Repository)
	}
	return fmt.Sprintf("## %s\n\n- %s %s.%s\n", r.Version, r.Binary, r.Version, link)
}

// Description fits a description to pub.dev's 60 to 180 characters: a
// longer one is cut at a word and ends with "...", a shorter one is an
// error, since padding it would say nothing.
func Description(s string) (string, error) {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) < 60 {
		return "", fmt.Errorf("pub.dev needs a description of 60 to 180 characters; the release description has %d", len(s))
	}
	if len(s) <= 180 {
		return s, nil
	}
	cut := strings.LastIndex(s[:177], " ")
	if cut < 60 {
		cut = 177
	}
	return strings.TrimRight(s[:cut], " ,.;:") + "...", nil
}
