using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The plugins of a stack.</summary>
public static class StackPlugins
{
    /// <summary>The stack's plugins other than its language and scaffold, in the order they apply: toolchain, structure, deploy.</summary>
    /// <param name="stack">The stack.</param>
    /// <returns>The plugin names.</returns>
    public static IReadOnlyList<string> Of(Stack stack)
    {
        ArgumentNullException.ThrowIfNull(stack);
        return [.. stack.Toolchain, .. stack.Structure, .. stack.Deploy];
    }
}
