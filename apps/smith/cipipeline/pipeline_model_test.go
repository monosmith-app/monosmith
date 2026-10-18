package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func npmAuth() protocol.Toolchain {
	return protocol.Toolchain{Name: "npm-registry-auth", Azure: &protocol.ActionStep{Uses: "npmAuthenticate@0", Inputs: []protocol.Input{{Name: "workingFile", Value: ".npmrc"}}}}
}

func TestAzureSignsInToPrivateRegistriesBeforeSetup(t *testing.T) {
	m := model()
	m.Authenticate = []protocol.Toolchain{npmAuth()}
	az := RenderAzure(m)
	auth := "  - task: npmAuthenticate@0\n    inputs:\n      workingFile: '.npmrc'\n"
	i, j := strings.Index(az, auth), strings.Index(az, "smith ci setup")
	if i < 0 || j < 0 || i > j {
		t.Fatalf("Azure: no sign-in before setup:\n%s", az)
	}
	// GitHub and GitLab have no such step: their credentials come from the
	// environment.
	if gh := RenderGitHub(m); strings.Contains(gh, "npmAuthenticate") {
		t.Errorf("GitHub got the Azure task:\n%s", gh)
	}
	if gl := RenderGitLab(m); strings.Contains(gl, "npmAuthenticate") || strings.Contains(gl, "npm-registry-auth") {
		t.Errorf("GitLab got the Azure task:\n%s", gl)
	}
}

func TestTheNativeJobSignsInToo(t *testing.T) {
	m := model()
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	m.Authenticate = []protocol.Toolchain{npmAuth()}
	if az := RenderAzure(m); strings.Count(az, "npmAuthenticate@0") != 2 {
		t.Errorf("Azure: the ci and native jobs both sign in:\n%s", az)
	}
}

func TestSetupGetsTheCredentialsInstallsNeed(t *testing.T) {
	m := model()
	m.Credentials = []protocol.Credential{{Env: "FEED_TOKEN", Setup: true}, {Env: "NODE_AUTH_TOKEN"}}
	gh, az := RenderGitHub(m), RenderAzure(m)
	if !strings.Contains(gh, "      - name: Setup\n        run: smith ci setup\n        env:\n          FEED_TOKEN: ${{ secrets.FEED_TOKEN }}\n      - name: Audit") {
		t.Errorf("GitHub setup lacks FEED_TOKEN only:\n%s", gh)
	}
	if !strings.Contains(az, "  - bash: smith ci setup\n    displayName: Setup\n    env:\n      FEED_TOKEN: $(FEED_TOKEN)\n  - bash: smith ci audit") {
		t.Errorf("Azure setup lacks FEED_TOKEN only:\n%s", az)
	}
	// The release keeps every credential.
	if !strings.Contains(gh, "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}") || !strings.Contains(az, "NODE_AUTH_TOKEN: $(NODE_AUTH_TOKEN)") {
		t.Error("the release lost a credential")
	}
}

func TestPostReleaseStepsAreMergedInOrderAndLeaveThePipelineYAMLAlone(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{PostRelease: []protocol.Step{{Name: "first", Command: []string{"a"}}}})
	c.Add(protocol.CIResult{PostRelease: []protocol.Step{{Name: "second", Command: []string{"b"}}}})
	if len(c.PostRelease) != 2 || c.PostRelease[0].Name != "first" || c.PostRelease[1].Name != "second" {
		t.Errorf("post-release steps %+v", c.PostRelease)
	}
	// smith ci release runs them, so the generated YAML is the same with or without.
	with := RenderGitHub(NewModel(c, "1.0.0", "main"))
	without := RenderGitHub(NewModel(Contributions{}, "1.0.0", "main"))
	if with != without {
		t.Error("a post-release step changed the generated GitHub workflow")
	}
}

func TestContributionsMergeSetupCredentialsAndSignIns(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "T"}}, Authenticate: []protocol.Toolchain{npmAuth()}})
	c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "T", Setup: true}}, Authenticate: []protocol.Toolchain{npmAuth()}})
	if len(c.Credentials) != 1 || !c.Credentials[0].Setup {
		t.Errorf("credentials: %+v", c.Credentials)
	}
	if len(c.Authenticate) != 1 {
		t.Errorf("sign-ins: %+v", c.Authenticate)
	}
	if m := NewModel(c, "1.0.0", "main"); len(m.Authenticate) != 1 || len(m.SetupCredentials()) != 1 {
		t.Errorf("model: %+v", m)
	}
}
