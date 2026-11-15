using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace Monosmith.Launcher.ReleaseDownload;

/// <summary>Reads smith's release files from GitHub or a mirror (SMITH_DOWNLOAD_URL).</summary>
internal static class ReleasesClient
{
    private const string GitHub = "https://github.com/monosmith-app/monosmith/releases/download";
    private static readonly Uri Api = new("https://api.github.com/repos/monosmith-app/monosmith/releases?per_page=100");
    private static readonly HttpClient Http = CreateClient();

    private static HttpClient CreateClient()
    {
        var client = new HttpClient { Timeout = TimeSpan.FromMinutes(5) };
        client.DefaultRequestHeaders.UserAgent.Add(new ProductInfoHeaderValue("monosmith-launcher", "1"));
        return client;
    }

    /// <summary>The address a release's files are under: SMITH_DOWNLOAD_URL (a mirror), else GitHub.</summary>
    /// <param name="env">The environment.</param>
    /// <returns>The base, without a trailing slash.</returns>
    internal static string DownloadBase(IReadOnlyDictionary<string, string> env) =>
        (env.TryGetValue("SMITH_DOWNLOAD_URL", out var mirror) && mirror.Length > 0 ? mirror : GitHub).TrimEnd('/');

    private static byte[] Get(Uri url, string? token = null)
    {
        using var request = new HttpRequestMessage(HttpMethod.Get, url);
        if (token is not null)
        {
            request.Headers.Accept.ParseAdd("application/vnd.github+json");
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", token);
        }
        HttpResponseMessage response;
        try
        {
            response = Http.Send(request);
        }
        catch (HttpRequestException error)
        {
            throw new HttpRequestException($"cannot reach {url.Host} ({error.Message}); behind a firewall, set SMITH_DOWNLOAD_URL to a mirror of smith's release archives", error);
        }
        using (response)
        {
            if (!response.IsSuccessStatusCode)
            {
                throw new HttpRequestException($"{url} answered {(int)response.StatusCode}");
            }
            using var body = new MemoryStream();
            response.Content.ReadAsStream().CopyTo(body);
            return body.ToArray();
        }
    }

    /// <summary>One file of a release: &lt;base&gt;/smith@&lt;version&gt;/&lt;name&gt;.</summary>
    /// <param name="env">The environment.</param>
    /// <param name="version">An exact version.</param>
    /// <param name="name">The file's name.</param>
    /// <returns>Its bytes.</returns>
    /// <exception cref="HttpRequestException">The file cannot be downloaded.</exception>
    internal static byte[] ReleaseFile(IReadOnlyDictionary<string, string> env, string version, string name) =>
        Get(new Uri($"{DownloadBase(env)}/smith%40{version}/{name}"));

    /// <summary>
    /// The newest stable smith: the mirror's latest.txt, or the newest smith@ release GitHub lists (GITHUB_TOKEN or
    /// GH_TOKEN, when set, raises GitHub's rate limit).
    /// </summary>
    /// <param name="env">The environment.</param>
    /// <returns>The version.</returns>
    /// <exception cref="InvalidOperationException">No release can be found.</exception>
    internal static string NewestRelease(IReadOnlyDictionary<string, string> env)
    {
        if (env.TryGetValue("SMITH_DOWNLOAD_URL", out var mirror) && mirror.Length > 0)
        {
            return Encoding.UTF8.GetString(Get(new Uri($"{DownloadBase(env)}/latest.txt"))).Trim();
        }
        var token = env.TryGetValue("GITHUB_TOKEN", out var t) && t.Length > 0 ? t : env.TryGetValue("GH_TOKEN", out var g) && g.Length > 0 ? g : null;
        using var releases = JsonDocument.Parse(Get(Api, token));
        foreach (var release in releases.RootElement.EnumerateArray())
        {
            var tag = release.TryGetProperty("tag_name", out var name) ? name.GetString() ?? "" : "";
            if (tag.StartsWith("smith@", StringComparison.Ordinal) && !IsTrue(release, "prerelease") && !IsTrue(release, "draft"))
            {
                return tag["smith@".Length..];
            }
        }
        throw new InvalidOperationException("GitHub lists no smith release");
    }

    private static bool IsTrue(JsonElement element, string property) =>
        element.TryGetProperty(property, out var value) && value.ValueKind == JsonValueKind.True;
}
