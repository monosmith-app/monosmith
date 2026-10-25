// A smith structure plugin in C# on MonoSmith.Plugin.Sdk: it adds a layout-check task to widget projects and makes
// their lint wait for it.
//
// Started by smith (SMITH_PLUGIN=1) it serves the plugin protocol on stdin and stdout. Started with `layout-check`,
// which is the command of the task it contributes, it checks the folder it runs in. The SDK frames the messages,
// derives the capabilities from the interfaces of the plugin (LayoutPlugin.cs) and keeps stdout for the protocol:
// anything meant for a human goes to stderr (Console.Error, or a plain Console.WriteLine once serving).
using MonoSmith.Plugin.Sdk;
using StructureLayout;

if (Environment.GetEnvironmentVariable("SMITH_PLUGIN") == "1")
{
    return await PluginServer.RunAsync(new LayoutPlugin());
}

if (args is ["layout-check"])
{
    var found = LayoutPlugin.FindJunkDrawers(".");
    if (found.Count > 0)
    {
        Console.Error.WriteLine($"folders named for nothing: {string.Join(", ", found)} (name them for what they do)");
        return 1;
    }

    Console.WriteLine("no junk-drawer folders");
    return 0;
}

Console.Error.WriteLine("usage: StructureLayout layout-check (smith runs this program as a plugin)");
return 1;
