package cachesetup

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
)

// fakeAz answers the az commands of the setup and records every command run.
type fakeAz struct{ ran []string }

func (f *fakeAz) run(_ string, argv ...string) (string, string, error) {
	line := CommandLine(argv)
	f.ran = append(f.ran, line)
	switch {
	case strings.HasPrefix(line, "az account list"):
		return `[{"name":"Prod","id":"sub-1","isDefault":true}]`, "", nil
	case strings.HasPrefix(line, "az storage account list"):
		return `[{"name":"acme","isHnsEnabled":false}]`, "", nil
	case strings.HasPrefix(line, "az storage container list"):
		return `[{"name":"cache"}]`, "", nil
	case strings.Contains(line, "blob list") && strings.Contains(line, "--prefix"):
		return `[]`, "", nil
	case strings.Contains(line, "blob list"):
		return `[{"name":"smith/"},{"name":"readme.txt"}]`, "", nil
	}
	return "", "unexpected " + line, errors.New("unexpected")
}

func ports(input string, az *fakeAz, out *bytes.Buffer, written *[]string) Ports {
	return Ports{
		Run:      az.run,
		LookPath: func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC) },
		WriteRemote: func(r string, _ bool) (string, error) {
			*written = append(*written, r)
			return "", nil
		},
		Dialogue: Terminal{In: bufio.NewReader(strings.NewReader(input)), Out: out, SecretFn: func(string, string, func() (string, error), []cachekey.KeyAction) error {
			return errors.New("no key was asked for")
		}},
	}
}

func TestAzureSetupListsShowsCommandsAndWritesAfterYes(t *testing.T) {
	az, out, written := &fakeAz{}, &bytes.Buffer{}, []string(nil)
	input := strings.Join([]string{"1", "", "1", "", "1", "", "1", "", "2", "", "1", "1", "1", "2", "2"}, "\n") + "\n"
	remote, err := Run(Facts{WorkspaceName: "Shop App", Providers: []string{"github", "azure"}}, ports(input, az, out, &written))
	if err != nil {
		t.Fatal(err)
	}
	if want := "azblob://acme/cache/smith/shop-app"; remote != want || len(written) != 1 || written[0] != want {
		t.Fatalf("remote %q written %v\n%s", remote, written, out)
	}
	for _, want := range []string{"smith will run this command. It only reads:", "az account list", "az storage account list --subscription sub-1", "--auth-mode login", "a prefix appears at the first write", "[cache]\n  remote = \"azblob://acme/cache/smith/shop-app\"", "GitHub Actions", "gh secret set SMITH_CACHE_KEY", "Azure Pipelines", "AZURE_STORAGE_SAS_TOKEN"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestNothingIsWrittenOrRunWithoutAYes(t *testing.T) {
	az, out, written := &fakeAz{}, &bytes.Buffer{}, []string(nil)
	// Azure; decline the subscription list (2), type the rest by hand; write: no (2); key: no; sas: no.
	input := strings.Join([]string{"1", "2", "sub", "2", "acme", "2", "cache", "2", "", "1", "2", "2", "2"}, "\n") + "\n"
	if _, err := Run(Facts{WorkspaceName: "shop"}, ports(input, az, out, &written)); err != nil {
		t.Fatal(err)
	}
	if len(az.ran) != 0 || len(written) != 0 {
		t.Fatalf("ran %v, wrote %v\n%s", az.ran, written, out)
	}
}

func TestAMissingCloudProgramFallsBackToTyping(t *testing.T) {
	az, out, written := &fakeAz{}, &bytes.Buffer{}, []string(nil)
	p := ports("2\nmy-bucket\n\n1\n1\n2\n", az, out, &written)
	p.LookPath = func(string) (string, error) { return "", errors.New("missing") }
	remote, err := Run(Facts{WorkspaceName: "shop"}, p)
	if err != nil || remote != "s3://my-bucket/shop" || !strings.Contains(out.String(), "`aws` is not installed here") {
		t.Fatalf("%q %v\n%s", remote, err, out)
	}
}

func TestSignInIsOfferedAsAChoiceNeverRunSilently(t *testing.T) {
	var ran []string
	out, written := &bytes.Buffer{}, []string(nil)
	attempt := 0
	p := ports("1\n\n2\nsub\n2\nacme\n2\nc\n2\n\n1\n2\n2\n", &fakeAz{}, out, &written)
	p.Run = func(_ string, argv ...string) (string, string, error) {
		ran = append(ran, CommandLine(argv))
		attempt++
		return "", "Please run 'az login' to setup account.", errors.New("exit 1")
	}
	if _, err := Run(Facts{WorkspaceName: "shop"}, p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Nobody is signed in to az.") || len(ran) != 1 {
		t.Fatalf("ran %v\n%s", ran, out)
	}
}

func TestHTTPSAddressIsCheckedAndCredentialsAreNamed(t *testing.T) {
	out, written := &bytes.Buffer{}, []string(nil)
	remote, err := Run(Facts{WorkspaceName: "shop", Providers: []string{"gitlab"}}, ports("3\nnot an address\nhttps://cache.acme.dev/smith\n1\n2\n", &fakeAz{}, out, &written))
	if err != nil || remote != "https://cache.acme.dev/smith" || !strings.Contains(out.String(), "SMITH_CACHE_TOKEN") || !strings.Contains(out.String(), "glab variable set SMITH_CACHE_KEY --masked") {
		t.Fatalf("%q %v\n%s", remote, err, out)
	}
}

func TestParsersReadTheCloudsOutput(t *testing.T) {
	items, err := AzureFolders("", "a", "c", "smith").Parse(`[{"name":"smith/ws1/"},{"name":"smith/x.txt"}]`)
	if err != nil || len(items) != 1 || items[0].Label != "ws1" || items[0].Value != "smith/ws1" {
		t.Fatalf("%v %v", items, err)
	}
	items, err = S3Folders("b", "").Parse(`{"CommonPrefixes":[{"Prefix":"smith/"}]}`)
	if err != nil || len(items) != 1 || items[0].Value != "smith" {
		t.Fatalf("%v %v", items, err)
	}
	if items, err := S3Folders("b", "").Parse(""); err != nil || items != nil {
		t.Fatalf("%v %v", items, err)
	}
	if _, err := AzureSubscriptions().Parse("not json"); err == nil {
		t.Fatal("want an error")
	}
	if line := CommandLine(GenerateSAS("", "a", "c", "login", time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC))); !strings.Contains(line, "--as-user") || !strings.Contains(line, "--expiry 2026-11-12T00:00Z") || strings.Contains(line, "--subscription") || !strings.Contains(line, "--permissions rwdcl") {
		t.Fatal(line)
	}
}

func TestWorkspaceSegmentAndValidation(t *testing.T) {
	for in, want := range map[string]string{"Shop App": "shop-app", "  ": "workspace", "a/b": "a-b"} {
		if got := WorkspaceSegment(in); got != want {
			t.Errorf("%q -> %q", in, got)
		}
	}
	if ValidateHTTPS("https://u:p@h/x") == "" || ValidateHTTPS("ftp://h") == "" || ValidateHTTPS("https://h/x") != "" {
		t.Fatal("address validation")
	}
}

func TestTheLoginWayExplainsItsSevenDaysAndBackAsksTheWayAgain(t *testing.T) {
	az, out, written := &fakeAz{}, &bytes.Buffer{}, []string(nil)
	// yes; default way (login); Back; default way (login); 7 days; skip the command.
	input := strings.Join([]string{"1", "", "2", "", "1", "2"}, "\n") + "\n"
	flow{Ports: ports(input, az, out, &written), azure: &azureState{}}.sasToken(sasTarget{subscription: "s", account: "acme", container: "cache"})
	text := out.String()
	if strings.Count(text, "How should the token be signed?") != 2 || strings.Count(text, "7 days is the only choice") != 2 {
		t.Fatalf("Back did not ask the way again:\n%s", text)
	}
	for _, want := range []string{"go back and sign with the account key", "renewing the CI secret every week"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
}
