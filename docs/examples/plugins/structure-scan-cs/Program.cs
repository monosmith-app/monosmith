// A smith structure plugin in C# on MonoSmith.Plugin.Sdk whose doctor walks the files of widget projects: it stops
// when smith cancels the request and it takes four scans at the same time. It only speaks the plugin protocol.
using System.Globalization;
using MonoSmith.Plugin.Sdk;
using StructureScan;

if (Environment.GetEnvironmentVariable("SMITH_PLUGIN") == "1")
{
    // WIDGET_SCAN_DELAY_MS is a pause before each file, which stands in for a slow disk so a test can cancel a scan in
    // the middle.
    var delay = double.TryParse(Environment.GetEnvironmentVariable("WIDGET_SCAN_DELAY_MS"), CultureInfo.InvariantCulture, out var milliseconds)
        ? TimeSpan.FromMilliseconds(Math.Max(milliseconds, 0))
        : TimeSpan.Zero;
    return await PluginServer.RunAsync(new ScanPlugin(delay));
}

Console.Error.WriteLine("StructureScan is a smith plugin: smith starts it and speaks the plugin protocol on its stdin and stdout");
return 1;
