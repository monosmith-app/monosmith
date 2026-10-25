using System.Text;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>Writing values the way the details of the Go checks do.</summary>
internal static class GoListMapper
{
    /// <summary>A list as Go's <c>%v</c> writes it, <c>[a b]</c>, which the details of the Go checks use.</summary>
    public static string List(IEnumerable<object> items) => "[" + string.Join(' ', items) + "]";

    /// <summary>A string as Go's <c>%q</c> writes it.</summary>
    public static string Quote(string text)
    {
        var output = new StringBuilder();
        GoJsonTextAlgorithm.AppendString(output, text, escapeHtml: false);
        return output.ToString();
    }
}
