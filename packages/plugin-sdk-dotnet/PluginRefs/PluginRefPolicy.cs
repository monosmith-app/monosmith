using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>References to other plugins: what a plugin requires or conflicts with.</summary>
public static class PluginRefs
{
    /// <summary>A requirement on the project's language plugin: <c>RequiresLanguage("node")</c>.</summary>
    /// <param name="name">The language plugin's name.</param>
    /// <returns>The reference.</returns>
    public static PluginRef RequiresLanguage(string name) => new() { Level = Role.Language, Name = name };

    /// <summary>A plugin that must not be in the same project.</summary>
    /// <param name="name">The plugin's name.</param>
    /// <returns>The reference.</returns>
    public static PluginRef ConflictsWith(string name) => new() { Name = name };

    /// <summary>Whether the reference names nothing: every member of a reference is optional, and at least one must be set.</summary>
    /// <param name="reference">The reference.</param>
    /// <returns>Whether it names nothing.</returns>
    public static bool IsEmpty(PluginRef reference)
    {
        ArgumentNullException.ThrowIfNull(reference);
        return reference.Level.IsEmpty && reference.Name.Length == 0 && reference.Slot.Length == 0;
    }

    /// <summary>How a message names it: <c>language plugin node</c>, <c>plugin providing test</c>.</summary>
    /// <param name="reference">The reference.</param>
    /// <returns>The description.</returns>
    public static string Describe(PluginRef reference)
    {
        ArgumentNullException.ThrowIfNull(reference);
        var parts = new List<string> { reference.Level.IsEmpty ? "plugin" : $"{reference.Level} plugin" };
        if (reference.Name.Length > 0)
        {
            parts.Add(reference.Name);
        }

        if (reference.Slot.Length > 0)
        {
            parts.Add($"providing {reference.Slot}");
        }

        return string.Join(' ', parts);
    }
}
