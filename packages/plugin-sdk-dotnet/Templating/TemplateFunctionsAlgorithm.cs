using System.Text;
using System.Text.RegularExpressions;

namespace MonoSmith.Plugin.Sdk;

/// <summary>A template function failed: its arguments are not what it takes.</summary>
internal sealed class TemplateFunctionError(string message) : Exception(message);

/// <summary>The functions of <c>text/template</c> that the supported subset has, and the way it prints values.</summary>
internal static partial class TemplateFunctionsAlgorithm
{
    /// <summary>A template function receives its evaluated arguments (the piped value last).</summary>
    public delegate JsonAny? TemplateFunction(IReadOnlyList<JsonAny?> arguments);

    /// <summary>The functions of <c>text/template</c> the subset has.</summary>
    public static IReadOnlyDictionary<string, TemplateFunction> Functions { get; } = new Dictionary<string, TemplateFunction>(StringComparer.Ordinal)
    {
        ["eq"] = Equal,
        ["ne"] = NotEqual,
        ["lt"] = Ordered(difference => difference < 0),
        ["le"] = Ordered(difference => difference <= 0),
        ["gt"] = Ordered(difference => difference > 0),
        ["ge"] = Ordered(difference => difference >= 0),
        ["not"] = arguments => new JsonBool(!IsTruthy(arguments[0])),
        ["and"] = arguments => arguments.FirstOrDefault(argument => !IsTruthy(argument), arguments[^1]),
        ["or"] = arguments => arguments.FirstOrDefault(argument => IsTruthy(argument), arguments[^1]),
        ["len"] = Length,
        ["index"] = Index,
        ["print"] = arguments => new JsonText(Sprint(arguments)),
        ["printf"] = Printf,
    };

    /// <summary>Whether the value is nil: absent or <c>null</c>.</summary>
    public static bool IsNil(JsonAny? value) => value is null or JsonNull;

    /// <summary>Go's truth: false, 0, nil, and an empty string, list or map are false; everything else is true.</summary>
    public static bool IsTruthy(JsonAny? value) => value switch
    {
        null or JsonNull => false,
        JsonBool boolean => boolean.Value,
        JsonNumber number => number.Real != 0,
        JsonText text => text.Value.Length > 0,
        JsonList list => list.Count > 0,
        JsonMap map => map.Count > 0,
        _ => true,
    };

    /// <summary>A value as <c>text/template</c> prints it.</summary>
    public static string FormatValue(JsonAny? value) => value switch
    {
        null or JsonNull => "<no value>",
        JsonBool boolean => boolean.Value ? "true" : "false",
        JsonList list => "[" + string.Join(' ', list.Select(FormatValue)) + "]",
        JsonMap map => "map[" + string.Join(' ', SortedKeys(map).Select(key => $"{key}:{FormatValue(map[key])}")) + "]",
        JsonText text => text.Value,
        _ => value.ToString() ?? string.Empty,
    };

    /// <summary>The keys of a map in the order Go sorts them (by their bytes).</summary>
    public static List<string> SortedKeys(JsonMap map)
    {
        var keys = map.Keys.ToList();
        keys.Sort(GoJsonTextAlgorithm.CompareUtf8);
        return keys;
    }

    private static string? ScalarKind(JsonAny? value) => value switch
    {
        JsonBool => "boolean",
        JsonText => "string",
        JsonNumber => "number",
        _ => null,
    };

    private static string Comparable(JsonAny? a, JsonAny? b)
    {
        var kind = ScalarKind(a);
        if (kind is null || ScalarKind(b) != kind)
        {
            throw new TemplateFunctionError("incompatible types for comparison");
        }

        return kind;
    }

    private static bool SameValue(JsonAny a, JsonAny b) => a switch
    {
        JsonNumber left => left.Real == ((JsonNumber)b).Real,
        JsonText left => left.Value == ((JsonText)b).Value,
        JsonBool left => left.Value == ((JsonBool)b).Value,
        _ => false,
    };

    private static TemplateFunction Ordered(Func<int, bool> test) =>
        arguments =>
        {
            if (arguments.Count != 2)
            {
                throw new TemplateFunctionError("wrong number of args: want 2");
            }

            var (a, b) = (arguments[0], arguments[1]);
            if (Comparable(a, b) == "boolean")
            {
                throw new TemplateFunctionError("invalid type for comparison");
            }

            var difference = a is JsonNumber left
                ? left.Real.CompareTo(((JsonNumber)b!).Real)
                : GoJsonTextAlgorithm.CompareUtf8(((JsonText)a!).Value, ((JsonText)b!).Value);
            return new JsonBool(test(difference));
        };

    private static JsonBool Equal(IReadOnlyList<JsonAny?> arguments)
    {
        if (arguments.Count < 2)
        {
            throw new TemplateFunctionError("missing argument for comparison");
        }

        var first = arguments[0];
        foreach (var other in arguments.Skip(1))
        {
            if (IsNil(first) || IsNil(other))
            {
                if (IsNil(first) && IsNil(other))
                {
                    return new JsonBool(true);
                }

                continue;
            }

            Comparable(first, other);
            if (SameValue(first!, other!))
            {
                return new JsonBool(true);
            }
        }

        return new JsonBool(false);
    }

    private static JsonBool NotEqual(IReadOnlyList<JsonAny?> arguments)
    {
        if (arguments.Count != 2)
        {
            throw new TemplateFunctionError("wrong number of args: want 2");
        }

        var (a, b) = (arguments[0], arguments[1]);
        if (IsNil(a) || IsNil(b))
        {
            return new JsonBool(!(IsNil(a) && IsNil(b)));
        }

        Comparable(a, b);
        return new JsonBool(!SameValue(a!, b!));
    }

    private static JsonNumber Length(IReadOnlyList<JsonAny?> arguments)
    {
        var value = arguments[0];
        return value switch
        {
            JsonText text => new JsonNumber(Encoding.UTF8.GetByteCount(text.Value)),
            JsonList list => new JsonNumber(list.Count),
            JsonMap map => new JsonNumber(map.Count),
            _ => throw new TemplateFunctionError($"len of {FormatValue(value)}"),
        };
    }

    private static JsonAny? Index(IReadOnlyList<JsonAny?> arguments)
    {
        var value = arguments[0];
        foreach (var key in arguments.Skip(1))
        {
            value = IndexInto(value, key);
        }

        return value;
    }

    private static JsonAny? IndexInto(JsonAny? current, JsonAny? key)
    {
        switch (current)
        {
            case JsonList list:
                if (key is not JsonNumber { Integer: { } position } || position < 0 || position >= list.Count)
                {
                    throw new TemplateFunctionError($"index out of range: {FormatValue(key)}");
                }

                return list[(int)position];
            case JsonMap map:
                return map.TryGetValue(FormatKey(key), out var member) ? member : null;
            default:
                throw new TemplateFunctionError($"can't index {FormatValue(current)}");
        }
    }

    private static string FormatKey(JsonAny? key) => key switch
    {
        JsonText text => text.Value,
        JsonBool boolean => boolean.Value ? "True" : "False",
        _ => FormatValue(key),
    };

    /// <summary>What Sprint puts between operands: a space when neither side is a string.</summary>
    private static string Sprint(IReadOnlyList<JsonAny?> arguments)
    {
        var output = new StringBuilder();
        for (var i = 0; i < arguments.Count; i++)
        {
            if (i > 0 && arguments[i] is not JsonText && arguments[i - 1] is not JsonText)
            {
                output.Append(' ');
            }

            output.Append(FormatValue(arguments[i]));
        }

        return output.ToString();
    }

    private static JsonText Printf(IReadOnlyList<JsonAny?> arguments)
    {
        if (arguments.Count == 0 || arguments[0] is not JsonText pattern)
        {
            throw new TemplateFunctionError("printf needs a format string");
        }

        var next = 1;
        var result = Verb().Replace(pattern.Value, match =>
        {
            var verb = match.Groups[1].Value;
            if (verb == "%")
            {
                return "%";
            }

            if (next >= arguments.Count)
            {
                throw new TemplateFunctionError($"missing argument for %{verb}");
            }

            var argument = arguments[next++];
            if (verb != "q")
            {
                return FormatValue(argument);
            }

            var quoted = new StringBuilder();
            GoJsonTextAlgorithm.AppendString(quoted, FormatValue(argument), escapeHtml: false);
            return quoted.ToString();
        });
        return new JsonText(result);
    }

    [GeneratedRegex("%([%sdvqt])", RegexOptions.CultureInvariant)]
    private static partial Regex Verb();
}
