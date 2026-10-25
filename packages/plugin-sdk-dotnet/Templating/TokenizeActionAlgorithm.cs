using System.Globalization;
using System.Text.RegularExpressions;

namespace MonoSmith.Plugin.Sdk;

/// <summary>What an element of an action is.</summary>
internal enum TokenKind
{
    Open,
    Close,
    Pipe,
    Field,
    String,
    Number,
    Word,
}

/// <summary>One element of an action.</summary>
/// <param name="Kind">What it is.</param>
/// <param name="Value">The literal of a string or number, or the name of a word.</param>
/// <param name="Root">For a field: whether it starts at the root (<c>$</c>).</param>
/// <param name="Path">For a field: the names after the dots.</param>
internal sealed record Token(TokenKind Kind, JsonAny? Value = null, bool Root = false, IReadOnlyList<string>? Path = null);

/// <summary>Splitting the text of a template action into tokens.</summary>
internal static partial class TokenizeActionAlgorithm
{
    // The elements of an action, in the order the alternatives are tried: space, brackets and pipe, a quoted or raw
    // string, a number, a field (.A.B, $, $.A, or a lone .), a word, and := or =, which are variables. Anything else
    // is the last alternative, so every character belongs to some match.
    [GeneratedRegex(
        """
        [ \t\r\n\f\v]+|(?<open>\()|(?<close>\))|(?<pipe>\|)|(?<string>"(?:[^"\\]|\\.)*")|(?<raw>`[^`]*`)|(?<number>-?[0-9]+(?:\.[0-9]+)?(?![A-Za-z0-9_.]))|(?<field>\$(?:\.[A-Za-z_][A-Za-z0-9_]*)*|(?:\.[A-Za-z_][A-Za-z0-9_]*)+|\.)|(?<word>[A-Za-z_][A-Za-z0-9_]*)|(?<variable>:?=)|(?<other>.)
        """,
        RegexOptions.Singleline | RegexOptions.CultureInvariant)]
    private static partial Regex Element();

    /// <summary>Splits the text of an action into tokens. <paramref name="fail"/> reports what cannot be a token.</summary>
    public static IReadOnlyList<Token> Tokenize(string body, Func<string, Exception> fail)
    {
        var tokens = new List<Token>();
        foreach (Match match in Element().Matches(body))
        {
            if (match.Groups["open"].Success)
            {
                tokens.Add(new Token(TokenKind.Open));
            }
            else if (match.Groups["close"].Success)
            {
                tokens.Add(new Token(TokenKind.Close));
            }
            else if (match.Groups["pipe"].Success)
            {
                tokens.Add(new Token(TokenKind.Pipe));
            }
            else if (match.Groups["string"].Success)
            {
                tokens.Add(new Token(TokenKind.String, Unquote(match.Groups["string"].Value, fail)));
            }
            else if (match.Groups["raw"].Success)
            {
                var raw = match.Groups["raw"].Value;
                tokens.Add(new Token(TokenKind.String, new JsonText(raw[1..^1])));
            }
            else if (match.Groups["number"].Success)
            {
                tokens.Add(new Token(TokenKind.Number, NumberOf(match.Groups["number"].Value)));
            }
            else if (match.Groups["field"].Success)
            {
                var text = match.Groups["field"].Value;
                var path = text.Split('.').Where(part => part is not ("" or "$")).ToList();
                tokens.Add(new Token(TokenKind.Field, Root: text.StartsWith('$'), Path: path));
            }
            else if (match.Groups["word"].Success)
            {
                tokens.Add(new Token(TokenKind.Word, new JsonText(match.Groups["word"].Value)));
            }
            else if (match.Groups["variable"].Success)
            {
                throw fail("variables are not supported by the SDK's templates");
            }
            else if (match.Groups["other"].Success)
            {
                throw fail($"unexpected character \"{match.Groups["other"].Value}\" in action");
            }
        }

        return tokens;
    }

    private static JsonNumber NumberOf(string text) =>
        text.Contains('.', StringComparison.Ordinal) || !long.TryParse(text, NumberStyles.AllowLeadingSign, CultureInfo.InvariantCulture, out var integer)
            ? new JsonNumber(double.Parse(text, CultureInfo.InvariantCulture))
            : new JsonNumber(integer);

    private static JsonAny Unquote(string literal, Func<string, Exception> fail)
    {
        try
        {
            return JsonAny.Parse(literal) as JsonText ?? throw fail($"invalid string {literal}");
        }
        catch (JsonParseError)
        {
            throw fail($"invalid string {literal}");
        }
    }
}
