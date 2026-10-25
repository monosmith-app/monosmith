// A smith language plugin for "widget" projects (a directory with a widget.json), in C# on MonoSmith.Plugin.Sdk.
// Started by smith (SMITH_PLUGIN=1) it serves the plugin protocol on stdin and stdout; started by a person, or by
// one of the tasks it contributes, it is a small command line tool.
using LanguageWidget;
using MonoSmith.Plugin.Sdk;

if (Environment.GetEnvironmentVariable("SMITH_PLUGIN") == "1")
{
    return await PluginServer.RunAsync(new WidgetPlugin());
}

try
{
    WidgetTool.Run(args);
    return 0;
}
catch (InvalidOperationException error)
{
    Console.Error.WriteLine(error.Message);
    return 1;
}
