using System.Text;
using System.Text.RegularExpressions;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Formatting the text of a JSON array the way Go indents it.</summary>
internal static partial class IndentRawJsonAlgorithm
{
    // A string literal (escapes included), a bracket, a comma or colon, or a run of any other characters (a number,
    // true, null).
    [GeneratedRegex("""
        "(?:[^"\\]|\\.)*"|[{}\[\],:]|[^\s{}\[\],:"]+
        """, RegexOptions.CultureInvariant)]
    private static partial Regex Token();

    /// <summary>
    /// Formats JSON text as Go's <c>json.Indent(dst, src, prefix, "  ")</c> does. The insignificant whitespace is
    /// dropped and every element goes on its own line, indented two spaces per level and starting with
    /// <paramref name="prefix"/>; an empty array or object stays <c>[]</c> or <c>{}</c>. Text that is not an array is
    /// returned as it is. The text of leaves (numbers, strings) is kept exactly.
    /// </summary>
    public static string Indent(string raw, string prefix)
    {
        if (!raw.StartsWith('['))
        {
            return raw;
        }

        var output = new StringBuilder();
        var depth = 0;
        var needIndent = false; // Right after an opening bracket: held back until it is known whether the container is empty.

        void NewLine() => output.Append('\n').Append(prefix).Append(string.Concat(Enumerable.Repeat("  ", depth)));

        foreach (Match match in Token().Matches(raw))
        {
            var token = match.Value;
            if (needIndent && token is not ("}" or "]"))
            {
                needIndent = false;
                depth++;
                NewLine();
            }

            switch (token)
            {
                case "{" or "[":
                    needIndent = true;
                    output.Append(token);
                    break;
                case "}" or "]":
                    if (needIndent)
                    {
                        needIndent = false;
                    }
                    else
                    {
                        depth--;
                        NewLine();
                    }

                    output.Append(token);
                    break;
                case ",":
                    output.Append(token);
                    NewLine();
                    break;
                case ":":
                    output.Append(": ");
                    break;
                default:
                    output.Append(token);
                    break;
            }
        }

        return output.ToString();
    }
}
