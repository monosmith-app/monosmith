package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/nugettool"
	"github.com/monosmith-app/monosmith/apps/smith/pypiwheels"
	"github.com/monosmith-app/monosmith/apps/smith/rubygems"
)

// RegistryProbe asks the registries a release publishes to whether they would
// accept it, with real authenticated calls that publish nothing. The
// endpoints are fields so tests can point them at fake registries.
type RegistryProbe struct {
	Client *http.Client
	// NPMRegistry, PyPIUpload, PyPIAPI, NuGetPush, NuGetIndex and RubyGemsHost
	// are the registries' endpoints. An empty PyPIAPI or NuGetIndex skips the
	// version lookup (another index's JSON API is not known).
	NPMRegistry, PyPIUpload, PyPIAPI, NuGetPush, NuGetIndex, RubyGemsHost string
	// NPMToken, PyPIUser, PyPIPassword, NuGetKey and RubyGemsKey are the
	// credentials; empty means the registry is not asked (the credential
	// check says it is missing).
	NPMToken, PyPIUser, PyPIPassword, NuGetKey, RubyGemsKey string
	// Published says whether the release journal already holds a package,
	// which makes a registry having its version expected, not suspicious.
	Published func(project, registry, name string) bool
}

// ProbeFromEnvironment reads the endpoints and credentials the publishing
// code reads, so the check asks what the publish will.
func ProbeFromEnvironment() *RegistryProbe {
	p := &RegistryProbe{
		Client:       &http.Client{Timeout: 30 * time.Second},
		NPMRegistry:  orDefault(os.Getenv("NPM_CONFIG_REGISTRY"), "https://registry.npmjs.org"),
		PyPIUpload:   orDefault(os.Getenv("TWINE_REPOSITORY_URL"), pypiwheels.DefaultUploadURL),
		PyPIAPI:      "https://pypi.org/pypi",
		NuGetPush:    orDefault(os.Getenv("NUGET_PUSH_URL"), nugettool.DefaultPushURL),
		NuGetIndex:   "https://api.nuget.org/v3-flatcontainer",
		RubyGemsHost: orDefault(os.Getenv("RUBYGEMS_HOST"), rubygems.DefaultHost),
		NPMToken:     orDefault(os.Getenv("NODE_AUTH_TOKEN"), os.Getenv("NPM_TOKEN")),
		NuGetKey:     os.Getenv("NUGET_API_KEY"),
		RubyGemsKey:  os.Getenv("GEM_HOST_API_KEY"),
	}
	if u, err := pypiwheels.UploaderFromEnvironment(); err == nil {
		p.PyPIUser, p.PyPIPassword = u.Username, u.Password
	}
	if os.Getenv("TWINE_REPOSITORY_URL") != "" {
		p.PyPIAPI = ""
	}
	if os.Getenv("NUGET_PUSH_URL") != "" {
		p.NuGetIndex = ""
	}
	return p
}

func orDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// Check asks a target's registry about each of its packages.
func (p *RegistryProbe) Check(t RegistryTarget) []RegistryFinding {
	switch t.Registry {
	case "npm":
		return p.npm(t)
	case "PyPI":
		return p.pypi(t)
	case "NuGet":
		return p.nuget(t)
	case "RubyGems":
		return p.rubygems(t)
	}
	return nil
}

func (p *RegistryProbe) finding(t RegistryTarget, blocking bool, fix, format string, args ...any) RegistryFinding {
	return RegistryFinding{Project: t.Project, Registry: t.Registry, Message: fmt.Sprintf(format, args...), Fix: fix, Blocking: blocking}
}

func (p *RegistryProbe) unreachable(t RegistryTarget, host string, err error) RegistryFinding {
	return p.finding(t, true, "check the network and the registry's status, or the URL it is configured with", "%s is unreachable: %v", host, err)
}

// get fetches a URL with headers; the body is capped.
func (p *RegistryProbe) get(rawURL string, header map[string]string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, nil, err
	}
	return p.do(req, header)
}

func (p *RegistryProbe) do(req *http.Request, header map[string]string) (int, []byte, error) {
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return resp.StatusCode, body, nil
}

func rejected(status int) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden
}

func (p *RegistryProbe) npm(t RegistryTarget) []RegistryFinding {
	if p.NPMToken == "" {
		return nil // npm's own login: only `npm publish` can tell
	}
	registry := strings.TrimSuffix(p.NPMRegistry, "/")
	auth := map[string]string{"Authorization": "Bearer " + p.NPMToken}
	status, body, err := p.get(registry+"/-/whoami", auth)
	if err != nil {
		return []RegistryFinding{p.unreachable(t, registry, err)}
	}
	if rejected(status) {
		return []RegistryFinding{p.finding(t, true, "create an npm token that can publish "+strings.Join(t.Names, " and ")+" and set NODE_AUTH_TOKEN", "npm refused NODE_AUTH_TOKEN (%d)", status)}
	}
	var who struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(body, &who)
	var out []RegistryFinding
	for _, name := range t.Names {
		status, body, err := p.get(registry+"/"+strings.ReplaceAll(name, "/", "%2f"), auth)
		switch {
		case err != nil:
			return append(out, p.unreachable(t, registry, err))
		case status == http.StatusNotFound:
			continue // a new package
		case rejected(status):
			out = append(out, p.finding(t, true, "use a token that can read and publish "+name, "the token cannot see %s (%d)", name, status))
			continue
		case status != http.StatusOK:
			out = append(out, p.finding(t, true, "retry, or check the registry's status", "asking for %s gave %d", name, status))
			continue
		}
		var doc struct {
			Versions    map[string]json.RawMessage `json:"versions"`
			Maintainers []struct {
				Name string `json:"name"`
			} `json:"maintainers"`
		}
		_ = json.Unmarshal(body, &doc)
		if who.Username != "" && len(doc.Maintainers) > 0 {
			ours := false
			for _, m := range doc.Maintainers {
				ours = ours || strings.EqualFold(m.Name, who.Username)
			}
			// A scoped package can belong to an organisation's team, which
			// the maintainer list does not show: only a bare name is certain.
			if !ours {
				out = append(out, p.finding(t, !strings.HasPrefix(name, "@"), "publish under a name "+who.Username+" maintains", "%s exists and %s is not a maintainer", name, who.Username))
			}
		}
		if _, taken := doc.Versions[t.Version]; taken {
			out = append(out, p.taken(t, name))
		}
	}
	return out
}

// taken says a registry has the version already: expected when the journal
// shows this release published it, else worth a look, since the content
// cannot be compared before it is built.
func (p *RegistryProbe) taken(t RegistryTarget, name string) RegistryFinding {
	if p.Published != nil && p.Published(t.Project, t.Registry, name) {
		return p.finding(t, false, "", "%s %s is already published by this release; it will be skipped", name, t.Version)
	}
	return p.finding(t, false, "if it is not this release's, publish a new version", "%s %s is already published and not in the release journal; it will be skipped as the same content", name, t.Version)
}

func (p *RegistryProbe) pypi(t RegistryTarget) []RegistryFinding {
	if p.PyPIPassword == "" {
		return nil
	}
	// An upload with no file: the index checks the credentials first and
	// then refuses the request for lacking one, so nothing is published.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField(":action", "file_upload")
	_ = mw.WriteField("protocol_version", "1")
	_ = mw.Close()
	req, err := http.NewRequest(http.MethodPost, p.PyPIUpload, &body)
	if err != nil {
		return []RegistryFinding{p.unreachable(t, p.PyPIUpload, err)}
	}
	req.SetBasicAuth(p.PyPIUser, p.PyPIPassword)
	status, _, err := p.do(req, map[string]string{"Content-Type": mw.FormDataContentType()})
	if err != nil {
		return []RegistryFinding{p.unreachable(t, p.PyPIUpload, err)}
	}
	if rejected(status) {
		return []RegistryFinding{p.finding(t, true, "create a PyPI API token that can upload "+strings.Join(t.Names, " and ")+" and set PYPI_TOKEN", "PyPI refused the credentials (%d)", status)}
	}
	if p.PyPIAPI == "" {
		return nil
	}
	version, err := pypiwheels.Version(t.Version)
	if err != nil {
		return []RegistryFinding{p.finding(t, true, "use a version PyPI can spell", "%v", err)}
	}
	var out []RegistryFinding
	for _, name := range t.Names {
		status, _, err := p.get(strings.TrimSuffix(p.PyPIAPI, "/")+"/"+url.PathEscape(name)+"/"+url.PathEscape(version)+"/json", nil)
		switch {
		case err != nil:
			return append(out, p.unreachable(t, p.PyPIAPI, err))
		case status == http.StatusOK:
			out = append(out, p.taken(t, name))
		}
	}
	return out
}

func (p *RegistryProbe) nuget(t RegistryTarget) []RegistryFinding {
	if p.NuGetKey == "" {
		return nil
	}
	// A push with no package: the feed checks the key first.
	req, err := http.NewRequest(http.MethodPut, p.NuGetPush, strings.NewReader(""))
	if err != nil {
		return []RegistryFinding{p.unreachable(t, p.NuGetPush, err)}
	}
	status, _, err := p.do(req, map[string]string{"X-NuGet-ApiKey": p.NuGetKey, "X-NuGet-Client-Version": "7.0.0"})
	if err != nil {
		return []RegistryFinding{p.unreachable(t, p.NuGetPush, err)}
	}
	if rejected(status) {
		return []RegistryFinding{p.finding(t, true, "create a NuGet API key that can push "+strings.Join(t.Names, " and ")+" and set NUGET_API_KEY", "NuGet refused the API key (%d)", status)}
	}
	if p.NuGetIndex == "" {
		return nil
	}
	var out []RegistryFinding
	for _, name := range t.Names {
		status, body, err := p.get(strings.TrimSuffix(p.NuGetIndex, "/")+"/"+url.PathEscape(strings.ToLower(name))+"/index.json", nil)
		if err != nil {
			return append(out, p.unreachable(t, p.NuGetIndex, err))
		}
		var doc struct {
			Versions []string `json:"versions"`
		}
		if status == http.StatusOK && json.Unmarshal(body, &doc) == nil {
			for _, v := range doc.Versions {
				if strings.EqualFold(v, t.Version) {
					out = append(out, p.taken(t, name))
				}
			}
		}
	}
	return out
}

func (p *RegistryProbe) rubygems(t RegistryTarget) []RegistryFinding {
	if p.RubyGemsKey == "" {
		return nil
	}
	host := strings.TrimSuffix(p.RubyGemsHost, "/")
	auth := map[string]string{"Authorization": p.RubyGemsKey}
	status, body, err := p.get(host+"/api/v1/gems.json", auth)
	if err != nil {
		return []RegistryFinding{p.unreachable(t, host, err)}
	}
	if status == http.StatusForbidden {
		// RubyGems answers 403 to a key that is valid but was not given the
		// index scope this listing needs; a push-only key is the recommended one.
		return []RegistryFinding{p.finding(t, false, "give the key the index_rubygems scope to have ownership and versions checked before publishing", "the RubyGems key cannot list its gems (403), so ownership and existing versions were not checked")}
	}
	if rejected(status) {
		return []RegistryFinding{p.finding(t, true, "create a RubyGems API key that can push "+strings.Join(t.Names, " and ")+" and set GEM_HOST_API_KEY", "RubyGems refused the API key (%d)", status)}
	}
	var owned []struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(body, &owned)
	version, err := rubygems.Version(t.Version)
	if err != nil {
		return []RegistryFinding{p.finding(t, true, "use a version RubyGems can spell", "%v", err)}
	}
	var out []RegistryFinding
	for _, name := range t.Names {
		status, _, err := p.get(host+"/api/v1/gems/"+url.PathEscape(name)+".json", auth)
		if err != nil {
			return append(out, p.unreachable(t, host, err))
		}
		if status != http.StatusOK {
			continue // a new gem
		}
		ours := false
		for _, g := range owned {
			ours = ours || g.Name == name
		}
		if !ours {
			out = append(out, p.finding(t, true, "publish under a name this key owns", "the gem %s exists and the key's owner does not own it", name))
			continue
		}
		status, body, err := p.get(host+"/api/v1/versions/"+url.PathEscape(name)+".json", auth)
		if err != nil {
			return append(out, p.unreachable(t, host, err))
		}
		var versions []struct {
			Number string `json:"number"`
		}
		if status == http.StatusOK && json.Unmarshal(body, &versions) == nil {
			for _, v := range versions {
				if v.Number == version {
					out = append(out, p.taken(t, name))
					break
				}
			}
		}
	}
	return out
}
