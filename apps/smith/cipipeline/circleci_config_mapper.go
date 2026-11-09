package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CircleCIPath is where the CircleCI configuration is written.
const CircleCIPath = ".circleci/config.yml"

// CircleCIImage is the image the jobs run on; the toolchains install onto it
// with the plugins' shell commands, as on GitLab CI.
const CircleCIImage = "debian:bookworm"

// RenderCircleCI renders the model as a CircleCI configuration: one `ci` job
// with the same phases as the other providers, a `native` job after it when a
// project is native, and one workflow that runs both on every branch.
//
// CircleCI has no pull request trigger (it builds branches), so the pull
// request steps test CIRCLE_PULL_REQUEST at run time and the release steps
// are limited to the release branch with a `when` condition.
func RenderCircleCI(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, `
version: 2.1

executors:
  smith:
    docker:
      - image: %s

commands:
  install-tools:
    description: Check out the code, install the toolchains and smith
    steps:
      - run:
          name: Install system packages
          command: apt-get update -qq && apt-get install -y -qq curl git openssh-client ca-certificates xz-utils >/dev/null
      - checkout
`, CircleCIImage)
	for _, t := range m.Toolchains {
		b.WriteString(circleciSetup(t))
	}
	b.WriteString(circleciRestoreStores(m.Stores))
	install := SmithInstallFor(m)
	command := "sh " + InstallScriptPath + " " + shellWord(m.SmithVersion)
	if !install.Script {
		command = install.ContainerCommand()
		if dir := install.ContainerPathDir(); dir != "" {
			command += "\n" + `echo 'export PATH="` + dir + `:$PATH"' >> "$BASH_ENV"`
		}
	}
	fmt.Fprintf(&b, `      - run:
          name: Install smith
          command: |
            %s
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$BASH_ENV"
`, strings.ReplaceAll(command, "\n", "\n            "))
	if len(m.Stores) > 0 {
		b.WriteString("  save-stores:\n    description: Keep the package download stores for the next run\n    steps:\n")
		b.WriteString(circleciSaveStores(m.Stores))
	}

	b.WriteString("\njobs:\n  ci:\n    executor: smith\n    steps:\n      - install-tools\n")
	b.WriteString(slot("      ", "before-setup"))
	b.WriteString(circleciRun("Setup", "smith ci setup"))
	b.WriteString(circleciRun("Audit", "smith ci audit"))
	b.WriteString(circleciRun("Verify", "smith ci verify"))
	b.WriteString(slot("      ", "after-verify"))
	if len(m.Stores) > 0 {
		b.WriteString("      - save-stores\n")
	}
	b.WriteString("      # The release preview comments on the pull request when the repository is on GitHub (GITHUB_TOKEN with\n")
	b.WriteString("      # pull-requests: write); on Bitbucket or elsewhere it stays in this job's log.\n")
	b.WriteString(circleciRun("Release preview", "if [ -n \"$CIRCLE_PULL_REQUEST\" ]; then smith release --dry-run --comment; fi"))
	var onRelease strings.Builder
	if m.Pack {
		var pack strings.Builder
		pack.WriteString(circleciRun("Pack", "smith ci pack"))
		pack.WriteString("      - store_artifacts:\n          path: dist/drop\n          destination: drop\n")
		if len(m.ChannelBranches) == 0 {
			onRelease.WriteString(pack.String())
		} else {
			// Packing stays on the release branch; the release also runs on the channels.
			b.WriteString(circleciWhen(circleciBranch(m.ReleaseBranch), pack.String()))
		}
	}
	onRelease.WriteString("      # Releasing pushes tags: the checkout key is read-only, so add a user key with write access\n")
	onRelease.WriteString("      # (Project Settings > SSH Keys > Add user key) for it to push with.\n")
	onRelease.WriteString(circleciRun("Release", "smith ci release"))
	b.WriteString(circleciWhen(m.branchCondition(circleciBranch, func(c []string) string { return "or: [ { " + strings.Join(c, " }, { ") + " } ]" }), onRelease.String()))
	if m.Native {
		b.WriteString(circleciNativeJob(m))
	}

	b.WriteString("\nworkflows:\n  ci:\n    jobs:\n      - ci\n")
	if circleciNative(m) {
		b.WriteString("      - native:\n          requires: [ci]\n")
	}
	if len(m.Credentials) > 0 {
		b.WriteString("\n# Secrets the release reads, as project environment variables (or in a context):\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "#   %s: %s\n", c.Env, c.Description)
		}
	}
	return b.String()
}

// circleciNative is true when the native job is written: a project is native
// and a runner is Linux, the only OS this renderer builds on.
func circleciNative(m Model) bool {
	return m.Native && len(circleciLinuxRunners(m)) > 0
}

// circleciLinuxRunners are the configured native runners that are Linux.
func circleciLinuxRunners(m Model) []string {
	var out []string
	for _, r := range m.NativeRunners {
		if isLinuxRunner(r) {
			out = append(out, r)
		}
	}
	return out
}

// isLinuxRunner says whether a GitHub runner label (ubuntu-latest) is Linux.
func isLinuxRunner(label string) bool {
	return strings.Contains(label, "ubuntu") || strings.Contains(label, "linux")
}

// circleciNativeJob builds the native projects after the ci job. The
// toolchains' install commands are written for Debian, so only the Linux
// runner is built here; the others are named in a comment.
func circleciNativeJob(m Model) string {
	var b strings.Builder
	var skipped []string
	for _, r := range m.NativeRunners {
		if !isLinuxRunner(r) {
			skipped = append(skipped, r)
		}
	}
	if len(skipped) > 0 {
		fmt.Fprintf(&b, "\n  # Not built on CircleCI: %s. The toolchains install with Debian commands, so this job\n  # builds on Linux only; add a job with a macOS or Windows executor that runs `smith ci native`.", strings.Join(skipped, ", "))
		b.WriteString("\n")
	}
	if !circleciNative(m) {
		return b.String()
	}
	b.WriteString("  native:\n    executor: smith\n    steps:\n      - install-tools\n")
	b.WriteString(circleciRun("Setup", "smith ci setup"))
	b.WriteString(circleciRun("Native", "smith ci native"))
	b.WriteString("      - store_artifacts:\n          path: dist/drop\n          destination: native-linux\n")
	return b.String()
}

// circleciBranch is the CircleCI condition that holds on one branch.
func circleciBranch(branch string) string {
	return fmt.Sprintf("equal: [ %s, << pipeline.git.branch >> ]", yamlWord(branch))
}

// circleciWhen wraps steps in a condition: one circleciBranch, or an "or" of
// them.
func circleciWhen(condition, steps string) string {
	return fmt.Sprintf("      - when:\n          condition:\n            %s\n          steps:\n%s", condition, indent(steps, "      "))
}

// circleciRun is a run step with a one-line command.
func circleciRun(name, command string) string {
	return fmt.Sprintf("      - run:\n          name: %s\n          command: %s\n", name, yamlValue(command))
}

// circleciSetup installs a toolchain with the plugin's shell command. Each
// CircleCI step is its own shell, and the commands (written for GitLab, where
// one shell runs the whole job) export PATH and friends for "the rest of the
// job", so what the command changed in the environment is appended to
// $BASH_ENV, which every later step reads.
func circleciSetup(t protocol.Toolchain) string {
	if t.Install == "" {
		return missing("      ", t)
	}
	script := strings.TrimSpace(t.Install)
	return fmt.Sprintf(`      - run:
          name: Install %s
          command: |
            export -p | sort > /tmp/smith-env-before
            %s
            export -p | sort > /tmp/smith-env-after
            { comm -13 /tmp/smith-env-before /tmp/smith-env-after | grep -Ev '^declare -x (_|PWD|OLDPWD|SHLVL)=' || true; } >> "$BASH_ENV"
`, t.Name, strings.ReplaceAll(script, "\n", "\n            "))
}

// circleciRestoreStores hashes each store's lockfiles into one file (cache
// keys take one file's checksum) and restores the store by that key, falling
// back to the latest cache of the same store when a lockfile changed.
func circleciRestoreStores(stores []protocol.PackageStore) string {
	var b strings.Builder
	for _, s := range stores {
		fmt.Fprintf(&b, `      - run:
          name: Hash the %[1]s lockfiles
          command: cat %[2]s > /tmp/smith-%[1]s.lock 2>/dev/null || true
      - restore_cache:
          name: Restore the %[1]s download store
          keys:
            - smith-%[1]s-v1-{{ arch }}-{{ checksum "/tmp/smith-%[1]s.lock" }}
            - smith-%[1]s-v1-{{ arch }}-
`, s.Name, strings.Join(s.Files, " "))
	}
	return b.String()
}

// circleciSaveStores saves each store under the key it was restored by.
func circleciSaveStores(stores []protocol.PackageStore) string {
	var b strings.Builder
	for _, s := range stores {
		fmt.Fprintf(&b, `      - save_cache:
          name: Cache the %[1]s download store
          key: smith-%[1]s-v1-{{ arch }}-{{ checksum "/tmp/smith-%[1]s.lock" }}
          paths:
            - %[2]s
`, s.Name, s.Path)
	}
	return b.String()
}

// shellWord is a value as one shell word: bare when plain, else single-quoted.
func shellWord(s string) string {
	if plainYAMLWord.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
