using System.Text;

namespace LanguageWidget;

/// <summary>The command line the plugin's tasks call.</summary>
public static class WidgetTool
{
    /// <summary>
    /// The tasks run this same program, so the example needs no other tool.
    /// <code>
    /// LanguageWidget.dll check            validate ./widget.json
    /// LanguageWidget.dll build &lt;out-dir&gt;  write &lt;out-dir&gt;/widget.txt
    /// </code>
    /// </summary>
    /// <param name="args">The command line.</param>
    public static void Run(string[] args)
    {
        if (args.Length == 0)
        {
            throw new InvalidOperationException("usage: LanguageWidget check | build <out-dir> (smith runs this program as a plugin)");
        }

        var manifest = WidgetManifest.Read(WidgetManifest.FileName);
        switch (args[0])
        {
            case "check":
                manifest.Validate();
                Console.WriteLine($"{manifest.Name} {manifest.Version} is valid");
                break;
            case "build":
                if (args.Length != 2)
                {
                    throw new InvalidOperationException("build takes the output directory");
                }

                Directory.CreateDirectory(args[1]);
                var output = Path.Combine(args[1], "widget.txt");
                File.WriteAllBytes(output, Encoding.UTF8.GetBytes($"{manifest.Name} {manifest.Version}\n"));
                Console.WriteLine($"wrote {output.Replace('\\', '/')}");
                break;
            default:
                throw new InvalidOperationException($"unknown command \"{args[0]}\"");
        }
    }
}
