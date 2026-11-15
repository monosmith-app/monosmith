using System.Net;
using System.Text;

using Monosmith.Launcher.SignatureCheck;
using Monosmith.Launcher.SmithRun;
using Monosmith.Launcher.VersionChoice;

namespace Monosmith.Launcher.Tests;

// The launcher against a release signed by smith release's own minisign code, with a test key.
public sealed class LauncherTests
{
    private const string Archive = "monosmith_1.2.3_linux_amd64.zip";
    private static readonly string PublicKey = File.ReadAllLines(FixturePath("test-key.pub"))[1];

    private static string FixturePath(string name) => Path.Combine(AppContext.BaseDirectory, "Fixtures", name);

    private static byte[] Fixture(string name) => File.ReadAllBytes(FixturePath(name));

    private static string Signature => Encoding.UTF8.GetString(Fixture("checksums.txt.minisig"));

    [Fact]
    public void AnArchiveListedInACorrectlySignedListIsTrusted() =>
        Assert.Null(VerifyChecksumsAlgorithm.CheckArchive(Fixture(Archive), Archive, Fixture("checksums.txt"), Signature, PublicKey));

    [Fact]
    public void AChangedListAnotherKeyOrAChangedArchiveIsNotTrusted()
    {
        var sums = Fixture("checksums.txt");
        Assert.Contains("signature", VerifyChecksumsAlgorithm.CheckArchive(Fixture(Archive), Archive, [.. sums, (byte)'x'], Signature, PublicKey), StringComparison.Ordinal);
        Assert.Contains("signature", VerifyChecksumsAlgorithm.CheckArchive(Fixture(Archive), Archive, sums, Signature), StringComparison.Ordinal);
        Assert.Contains("does not match its checksum", VerifyChecksumsAlgorithm.CheckArchive("other"u8.ToArray(), Archive, sums, Signature, PublicKey), StringComparison.Ordinal);
        Assert.Contains("does not list", VerifyChecksumsAlgorithm.CheckArchive(Fixture(Archive), "monosmith_1.2.3_darwin_arm64.zip", sums, Signature, PublicKey), StringComparison.Ordinal);
    }

    [Fact]
    public void ADigestIsFoundByTheFileNameOfItsLine()
    {
        Assert.Equal("cd34", VerifyChecksumsAlgorithm.ListedDigest("ab12  one.zip\ncd34  two.zip\n", "two.zip"));
        Assert.Null(VerifyChecksumsAlgorithm.ListedDigest("ab12  one.zip\n", "three.zip"));
    }

    [Fact]
    public void TheVersionComesFromSmithVersionThenTheWorkspacePin()
    {
        var root = Directory.CreateTempSubdirectory().FullName;
        File.WriteAllText(Path.Combine(root, "smith.toml"), "[ci]\nsmith = \"9.9.9\"\n[workspace]\nsmith = \"2.50.0\" # pinned\n");
        var deep = Directory.CreateDirectory(Path.Combine(root, "apps", "web")).FullName;
        Assert.Equal("2.50.0", ChooseVersionPolicy.Choose(new Dictionary<string, string>(), deep));
        Assert.Equal("3.0.0-beta.1", ChooseVersionPolicy.Choose(new Dictionary<string, string> { ["SMITH_VERSION"] = "3.0.0-beta.1" }, deep));
        Assert.Null(ChooseVersionPolicy.Choose(new Dictionary<string, string> { ["SMITH_VERSION"] = "latest" }, deep));
        Assert.Null(ChooseVersionPolicy.PinIn("[workspace]\nsmith = \"latest\"\n"));
        Assert.Throws<ArgumentException>(() => ChooseVersionPolicy.Choose(new Dictionary<string, string> { ["SMITH_VERSION"] = "^3" }, deep));
        Assert.True(ChooseVersionPolicy.Compare("3.0.0-beta.1", "3.0.0") < 0);
        Assert.True(ChooseVersionPolicy.Compare("2.10.0", "2.9.0") > 0);
    }

    [Fact]
    public void TheNewestReleaseIsDownloadedOnceCheckedAndKept()
    {
        using var mirror = new Mirror();
        var home = Directory.CreateTempSubdirectory().FullName;
        var env = new Dictionary<string, string> { ["SMITH_DOWNLOAD_URL"] = mirror.Base, ["SMITH_HOME"] = home };
        var program = RunSmithUseCase.ResolveProgram(env, home, _ => { }, "linux", "amd64", PublicKey);
        Assert.Equal(Path.Combine(home, "versions", "1.2.3", "smith"), program);
        Assert.Equal("#!/bin/sh\necho smith 1.2.3\n", File.ReadAllText(program));
        mirror.Requests.Clear();
        Assert.Equal(program, RunSmithUseCase.ResolveProgram(env, home, _ => { }, "linux", "amd64", PublicKey));
        Assert.Empty(mirror.Requests);
    }

    [Fact]
    public void AReleaseSignedWithAnotherKeyIsNotInstalled()
    {
        using var mirror = new Mirror();
        var home = Directory.CreateTempSubdirectory().FullName;
        var env = new Dictionary<string, string> { ["SMITH_DOWNLOAD_URL"] = mirror.Base, ["SMITH_HOME"] = home, ["SMITH_VERSION"] = "1.2.3" };
        var error = Assert.Throws<InvalidOperationException>(() => RunSmithUseCase.ResolveProgram(env, home, _ => { }, "linux", "amd64"));
        Assert.Contains("was not installed", error.Message, StringComparison.Ordinal);
        Assert.False(Directory.Exists(Path.Combine(home, "versions", "1.2.3")));
    }

    // A mirror of one release (SMITH_DOWNLOAD_URL), serving the files smith release attaches.
    private sealed class Mirror : IDisposable
    {
        private readonly HttpListener listener = new();
        private readonly Dictionary<string, byte[]> files = new(StringComparer.Ordinal) { ["/latest.txt"] = "1.2.3\n"u8.ToArray() };

        public Mirror()
        {
            foreach (var name in new[] { Archive, "checksums.txt", "checksums.txt.minisig" })
            {
                files[$"/smith%401.2.3/{name}"] = Fixture(name);
            }
            var port = new Random().Next(20000, 60000);
            Base = $"http://127.0.0.1:{port}";
            listener.Prefixes.Add(Base + "/");
            listener.Start();
            _ = Task.Run(Serve);
        }

        public string Base { get; }

        public List<string> Requests { get; } = [];

        private async Task Serve()
        {
            while (listener.IsListening)
            {
                HttpListenerContext context;
                try
                {
                    context = await listener.GetContextAsync().ConfigureAwait(false);
                }
                catch (HttpListenerException)
                {
                    return;
                }
                catch (ObjectDisposedException)
                {
                    return;
                }
                var path = context.Request.RawUrl ?? "";
                lock (Requests)
                {
                    Requests.Add(path);
                }
                var found = files.TryGetValue(path, out var body);
                context.Response.StatusCode = found ? 200 : 404;
                if (found)
                {
                    await context.Response.OutputStream.WriteAsync(body).ConfigureAwait(false);
                }
                context.Response.Close();
            }
        }

        public void Dispose() => listener.Close();
    }
}
