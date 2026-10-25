namespace MonoSmith.Plugin.Sdk;

/// <summary>A template that cannot be parsed or rendered: bad syntax, a field the data does not have, a construct outside the supported subset.</summary>
public sealed class TemplateError : Exception
{
    /// <summary>An error with no message.</summary>
    public TemplateError()
    {
    }

    /// <summary>An error with a message.</summary>
    /// <param name="message">The text of the error.</param>
    public TemplateError(string message)
        : base(message)
    {
    }

    /// <summary>An error with a message and a cause.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public TemplateError(string message, Exception innerException)
        : base(message, innerException)
    {
    }

    /// <summary>An error in the template with a name.</summary>
    /// <param name="name">The template's name: its target path, for a file.</param>
    /// <param name="message">What is wrong.</param>
    public TemplateError(string name, string message)
        : base($"template {name}: {message}")
    {
    }
}
