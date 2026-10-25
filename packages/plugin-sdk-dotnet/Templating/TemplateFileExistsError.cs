namespace MonoSmith.Plugin.Sdk;

/// <summary>A target file already exists: templates are written for a new project and never overwrite what is there.</summary>
public sealed class TemplateFileExistsError : IOException
{
    /// <summary>An error with no path.</summary>
    public TemplateFileExistsError()
        : this(string.Empty)
    {
    }

    /// <summary>An error for an existing file.</summary>
    /// <param name="path">The existing file, relative to the root.</param>
    public TemplateFileExistsError(string path)
        : base($"{path}: file already exists")
    {
        TargetPath = path;
    }

    /// <summary>An error for an existing file, with a cause.</summary>
    /// <param name="path">The existing file, relative to the root.</param>
    /// <param name="innerException">The cause.</param>
    public TemplateFileExistsError(string path, Exception innerException)
        : base($"{path}: file already exists", innerException)
    {
        TargetPath = path;
    }

    /// <summary>The existing file, relative to the root.</summary>
    public string TargetPath { get; }
}
