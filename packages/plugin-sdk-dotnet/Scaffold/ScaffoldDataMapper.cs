using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The data a project template can use.</summary>
public static class ScaffoldData
{
    /// <summary>What a project template can say, from the <c>generate</c> params.</summary>
    /// <remarks>
    /// <list type="table">
    /// <listheader><term>Field</term><description>Value</description></listheader>
    /// <item><term><c>Dir</c></term><description>the project's directory, workspace-relative</description></item>
    /// <item><term><c>Name</c></term><description>the project's name</description></item>
    /// <item><term><c>Kind</c></term><description><c>app</c>, <c>lib</c> or <c>package</c></description></item>
    /// <item><term><c>Stack</c></term><description>the stack to record in the manifest (may be empty)</description></item>
    /// <item><term><c>Workspace</c></term><description>the workspace's name</description></item>
    /// <item><term><c>Scope</c></term><description>the npm scope, with its <c>@</c> (may be empty)</description></item>
    /// <item><term><c>ModulePrefix</c></term><description>the Go module path prefix (may be empty)</description></item>
    /// </list>
    /// </remarks>
    /// <param name="parameters">The params of <c>generate</c>.</param>
    /// <returns>The fields.</returns>
    public static JsonMap From(GenerateParams parameters)
    {
        ArgumentNullException.ThrowIfNull(parameters);
        return new JsonMap
        {
            ["Dir"] = parameters.Directory,
            ["Name"] = parameters.Name,
            ["Kind"] = parameters.Kind.Value,
            ["Stack"] = parameters.Stack,
            ["Workspace"] = parameters.Workspace.Name,
            ["Scope"] = parameters.Workspace.Scope,
            ["ModulePrefix"] = parameters.Workspace.ModulePrefix,
        };
    }
}
