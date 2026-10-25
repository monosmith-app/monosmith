namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>An empty directory for a test to act as the workspace root, deleted when the test is done.</summary>
internal sealed class TempWorkspace : IDisposable
{
    public TempWorkspace() => Root = Directory.CreateTempSubdirectory("smith-sdk-").FullName;

    public string Root { get; }

    public string Path(string relative) => System.IO.Path.Combine(Root, relative.Replace('/', System.IO.Path.DirectorySeparatorChar));

    public string Read(string relative) => File.ReadAllText(Path(relative));

    public void Write(string relative, string text)
    {
        Directory.CreateDirectory(System.IO.Path.GetDirectoryName(Path(relative))!);
        File.WriteAllText(Path(relative), text);
    }

    public bool Exists(string relative) => File.Exists(Path(relative));

    /// <summary>Deletes a directory tree, read-only files (a git repository's objects) included. What cannot go is left to the system's temporary folder cleanup.</summary>
    public static void DeleteQuietly(string directory)
    {
        try
        {
            foreach (var file in Directory.EnumerateFiles(directory, "*", SearchOption.AllDirectories))
            {
                File.SetAttributes(file, FileAttributes.Normal);
            }

            Directory.Delete(directory, recursive: true);
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException)
        {
            // A process that was just stopped may still hold a file.
        }
    }

    public void Dispose() => DeleteQuietly(Root);
}
