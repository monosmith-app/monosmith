using System.Globalization;
using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Writing JSON text the way Go's <c>encoding/json</c> writes it. <c>System.Text.Json</c> cannot: it escapes
/// every non-ASCII letter by default, and the relaxed encoder leaves <c>&lt;</c>, <c>&gt;</c> and <c>&amp;</c> alone,
/// where Go escapes exactly those and leaves the letters.
/// </summary>
public static class GoJsonTextAlgorithm
{
    private const string Hex = "0123456789abcdef";

    /// <summary>Appends a string with its quotes, escaped as Go escapes it.</summary>
    /// <param name="output">Where to append.</param>
    /// <param name="text">The text. A lone surrogate, which has no UTF-8 form, is written as U+FFFD.</param>
    public static void AppendString(StringBuilder output, string text)
    {
        ArgumentNullException.ThrowIfNull(output);
        ArgumentNullException.ThrowIfNull(text);
        output.Append('"');
        for (var i = 0; i < text.Length; i++)
        {
            var c = text[i];
            switch (c)
            {
                case '"':
                    output.Append("\\\"");
                    break;
                case '\\':
                    output.Append("\\\\");
                    break;
                case '\b':
                    output.Append("\\b");
                    break;
                case '\f':
                    output.Append("\\f");
                    break;
                case '\n':
                    output.Append("\\n");
                    break;
                case '\r':
                    output.Append("\\r");
                    break;
                case '\t':
                    output.Append("\\t");
                    break;
                case '<' or '>' or '&' or '\x2028' or '\x2029':
                    AppendUnicodeEscape(output, c);
                    break;
                case < ' ':
                    AppendUnicodeEscape(output, c);
                    break;
                default:
                    if (char.IsHighSurrogate(c) && i + 1 < text.Length && char.IsLowSurrogate(text[i + 1]))
                    {
                        output.Append(c).Append(text[i + 1]);
                        i++;
                    }
                    else if (char.IsSurrogate(c))
                    {
                        output.Append('\xFFFD');
                    }
                    else
                    {
                        output.Append(c);
                    }

                    break;
            }
        }

        output.Append('"');
    }

    /// <summary>A number as Go writes a float64: the shortest form that reads back as the same number.</summary>
    /// <param name="value">A finite number.</param>
    /// <returns>The text: digits for the usual range, <c>1e+21</c> and <c>1e-7</c> beyond it.</returns>
    public static string FormatNumber(double value)
    {
        if (!double.IsFinite(value))
        {
            throw new ArgumentOutOfRangeException(nameof(value), value, "a JSON number is finite");
        }

        if (value == 0)
        {
            return double.IsNegative(value) ? "-0" : "0";
        }

        var magnitude = Math.Abs(value);
        var (digits, exponent) = ShortestDigits(magnitude);
        var sign = value < 0 ? "-" : string.Empty;
        if (magnitude < 1e-6 || magnitude >= 1e21)
        {
            var mantissa = digits.Length > 1 ? $"{digits[0]}.{digits[1..]}" : digits;
            var unsignedExponent = Math.Abs(exponent).ToString("00", CultureInfo.InvariantCulture);
            // Go cleans e-09 up to e-9 and leaves e+21 and e-10 as they are.
            if (exponent < 0 && unsignedExponent[0] == '0')
            {
                unsignedExponent = unsignedExponent[1..];
            }

            return $"{sign}{mantissa}e{(exponent < 0 ? '-' : '+')}{unsignedExponent}";
        }

        if (exponent >= 0)
        {
            return digits.Length <= exponent + 1
                ? sign + digits + new string('0', exponent + 1 - digits.Length)
                : $"{sign}{digits[..(exponent + 1)]}.{digits[(exponent + 1)..]}";
        }

        return $"{sign}0.{new string('0', -exponent - 1)}{digits}";
    }

    /// <summary>
    /// Compares two strings by the bytes of their UTF-8 form, the order Go sorts the keys of a map in. It is the
    /// order of the code points, which differs from the order of UTF-16 units where a surrogate pair meets a
    /// character above U+D7FF.
    /// </summary>
    /// <param name="left">A string.</param>
    /// <param name="right">Another string.</param>
    /// <returns>Negative, zero or positive, as <see cref="string.CompareOrdinal(string, string)"/> returns.</returns>
    public static int CompareUtf8(string left, string right)
    {
        ArgumentNullException.ThrowIfNull(left);
        ArgumentNullException.ThrowIfNull(right);
        var i = 0;
        var j = 0;
        while (i < left.Length && j < right.Length)
        {
            var a = NextCodePoint(left, ref i);
            var b = NextCodePoint(right, ref j);
            if (a != b)
            {
                return a < b ? -1 : 1;
            }
        }

        return (left.Length - i).CompareTo(right.Length - j);
    }

    private static int NextCodePoint(string text, ref int index)
    {
        var c = text[index++];
        if (char.IsHighSurrogate(c) && index < text.Length && char.IsLowSurrogate(text[index]))
        {
            return char.ConvertToUtf32(c, text[index++]);
        }

        return char.IsSurrogate(c) ? 0xFFFD : c;
    }

    private static void AppendUnicodeEscape(StringBuilder output, char c) =>
        output.Append("\\u").Append(Hex[(c >> 12) & 0xF]).Append(Hex[(c >> 8) & 0xF]).Append(Hex[(c >> 4) & 0xF]).Append(Hex[c & 0xF]);

    private static (string Digits, int Exponent) ShortestDigits(double magnitude)
    {
        var text = magnitude.ToString("R", CultureInfo.InvariantCulture);
        var exponent = 0;
        var e = text.IndexOf('E', StringComparison.Ordinal);
        if (e >= 0)
        {
            exponent = int.Parse(text[(e + 1)..], CultureInfo.InvariantCulture);
            text = text[..e];
        }

        var dot = text.IndexOf('.', StringComparison.Ordinal);
        var integerLength = dot < 0 ? text.Length : dot;
        var digits = dot < 0 ? text : text.Remove(dot, 1);
        var scientific = integerLength - 1 + exponent;
        var leading = 0;
        while (leading < digits.Length - 1 && digits[leading] == '0')
        {
            leading++;
        }

        digits = digits[leading..].TrimEnd('0');
        return (digits.Length == 0 ? "0" : digits, scientific - leading);
    }
}
