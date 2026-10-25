namespace MonoSmith.Plugin.Sdk.Tests.PluginRoles;

public sealed class PluginRolesTests
{
    [Fact]
    public void Knows_the_roles_of_the_protocol_in_the_order_of_a_make_up()
    {
        Assert.Equal(["language", "toolchain", "scaffold", "structure", "deploy"], Sdk.PluginRoles.All.Select(role => role.Value));
        Assert.True(Sdk.PluginRoles.IsKnown(Role.Deploy));
        Assert.False(Sdk.PluginRoles.IsKnown(new Role("nonsense")));
    }

    [Fact]
    public void Says_which_levels_are_exclusive_and_which_additive()
    {
        Assert.True(Sdk.PluginRoles.IsExclusive(Role.Language));
        Assert.True(Sdk.PluginRoles.IsExclusive(Role.Scaffold));
        Assert.False(Sdk.PluginRoles.IsExclusive(Role.Toolchain));
        Assert.True(Sdk.PluginRoles.IsAdditive(Role.Structure));
        Assert.True(Sdk.PluginRoles.IsAdditive(Role.Deploy));
        Assert.False(Sdk.PluginRoles.IsAdditive(Role.Language));
        Assert.False(Sdk.PluginRoles.IsAdditive(new Role("nonsense")));
    }

    [Fact]
    public void Gives_a_plugin_that_declares_no_roles_the_default_of_its_kind()
    {
        Assert.Equal(["language", "toolchain", "scaffold"], Sdk.PluginRoles.DefaultFor(PluginKind.Language).Select(role => role.Value));
        Assert.Equal(["structure"], Sdk.PluginRoles.DefaultFor(PluginKind.Convention).Select(role => role.Value));
        Assert.Equal(["structure"], Sdk.PluginRoles.DefaultFor(new PluginKind("weird")).Select(role => role.Value));
    }

    [Fact]
    public void Reads_the_roles_a_plugin_declared_or_falls_back_to_the_default()
    {
        var declared = new InitializeResult { Plugin = new Peer { Name = "p", Version = "1" }, Kind = PluginKind.Language, Capabilities = new Capabilities(), Roles = [Role.Toolchain] };
        Assert.Equal([Role.Toolchain], Sdk.PluginRoles.Effective(declared));
        Assert.Equal(["structure"], Sdk.PluginRoles.Effective(declared with { Roles = [], Kind = PluginKind.Convention }).Select(role => role.Value));
    }
}
