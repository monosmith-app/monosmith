package cachepipeline

import "testing"

func TestParseAzureRemote(t *testing.T) {
	want := AzureRepo{Organization: "acme", Project: "Shop App", Repository: "shop"}
	for _, raw := range []string{
		"https://dev.azure.com/acme/Shop%20App/_git/shop",
		"https://acme@dev.azure.com/acme/Shop%20App/_git/shop\n",
		"https://user:pat@dev.azure.com/acme/Shop%20App/_git/shop",
		"git@ssh.dev.azure.com:v3/acme/Shop%20App/shop",
		"ssh://git@ssh.dev.azure.com/v3/acme/Shop App/shop",
		"https://acme.visualstudio.com/Shop%20App/_git/shop",
		"https://acme.visualstudio.com/DefaultCollection/Shop%20App/_git/shop",
		"acme@vs-ssh.visualstudio.com:v3/acme/Shop%20App/shop",
	} {
		if got, ok := ParseAzureRemote(raw); !ok || got != want {
			t.Errorf("%q -> %+v %v", raw, got, ok)
		}
	}
	if got, ok := ParseAzureRemote("https://dev.azure.com/acme/_git/shop"); !ok || got.Project != "shop" || got.Repository != "shop" {
		t.Errorf("a project named like the repository: %+v %v", got, ok)
	}
	for _, raw := range []string{"", "https://github.com/acme/shop.git", "git@github.com:acme/shop.git", "https://dev.azure.com/acme", "/home/me/shop"} {
		if got, ok := ParseAzureRemote(raw); ok {
			t.Errorf("%q -> %+v, want no", raw, got)
		}
	}
}

func TestClassifyTheFailuresOfAz(t *testing.T) {
	for text, want := range map[string]Failure{
		"ERROR: Please run 'az login' to setup account.": FailureSignIn,
		"Before you can run Azure DevOps commands, you need to run the login command(az login if using AAD/MSA identity else az devops login)": FailureSignIn,
		"TF400813: The user is not authorized":                              FailureSignIn,
		"ERROR: 'pipelines' is misspelled or not recognized by the system.": FailureExtension,
		"something else": FailureOther,
	} {
		if got := Classify(text); got != want {
			t.Errorf("%q -> %v, want %v", text, got, want)
		}
	}
}
