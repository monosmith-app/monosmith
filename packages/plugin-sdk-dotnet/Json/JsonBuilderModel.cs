using System.Globalization;
using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Writes JSON text the way Go writes it: compact, with <c>&lt;</c>, <c>&gt;</c>, <c>&amp;</c>, U+2028 and U+2029
/// escaped and the other characters left as they are. The members of an object come in the order they are
/// written, so a generated message writes them in the order of the schema; a map is written with
/// <see cref="Any"/>, which sorts its keys.
/// </summary>
public sealed class JsonBuilder
{
    private readonly StringBuilder _text = new();
    private readonly List<ContainerState> _containers = [];

    /// <summary>Starts an object.</summary>
    /// <returns>This builder.</returns>
    public JsonBuilder BeginObject()
    {
        BeforeValue();
        _text.Append('{');
        _containers.Add(new ContainerState(IsObject: true));
        return this;
    }

    /// <summary>Ends the object that was started last.</summary>
    /// <returns>This builder.</returns>
    public JsonBuilder EndObject()
    {
        _containers.RemoveAt(_containers.Count - 1);
        _text.Append('}');
        return this;
    }

    /// <summary>Starts an array.</summary>
    /// <returns>This builder.</returns>
    public JsonBuilder BeginArray()
    {
        BeforeValue();
        _text.Append('[');
        _containers.Add(new ContainerState(IsObject: false));
        return this;
    }

    /// <summary>Ends the array that was started last.</summary>
    /// <returns>This builder.</returns>
    public JsonBuilder EndArray()
    {
        _containers.RemoveAt(_containers.Count - 1);
        _text.Append(']');
        return this;
    }

    /// <summary>Writes the name of the next member of the object being written.</summary>
    /// <param name="name">The member's name.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Property(string name)
    {
        var state = _containers[^1];
        if (state.Count > 0)
        {
            _text.Append(',');
        }

        state.Count++;
        GoJsonTextAlgorithm.AppendString(_text, name);
        _text.Append(':');
        return this;
    }

    /// <summary>Writes a string.</summary>
    /// <param name="value">The text; <c>null</c> is written as the empty string.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder String(string? value)
    {
        BeforeValue();
        GoJsonTextAlgorithm.AppendString(_text, value ?? string.Empty);
        return this;
    }

    /// <summary>Writes an integer.</summary>
    /// <param name="value">The integer.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Integer(long value)
    {
        BeforeValue();
        _text.Append(value.ToString(CultureInfo.InvariantCulture));
        return this;
    }

    /// <summary>Writes a number.</summary>
    /// <param name="value">A finite number; anything else cannot be written as JSON.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Number(double value)
    {
        if (!double.IsFinite(value))
        {
            throw new JsonEncodeError($"{value.ToString(CultureInfo.InvariantCulture)} cannot be written as JSON");
        }

        BeforeValue();
        _text.Append(GoJsonTextAlgorithm.FormatNumber(value));
        return this;
    }

    /// <summary>Writes a boolean.</summary>
    /// <param name="value">The boolean.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Boolean(bool value)
    {
        BeforeValue();
        _text.Append(value ? "true" : "false");
        return this;
    }

    /// <summary>Writes <c>null</c>.</summary>
    /// <returns>This builder.</returns>
    public JsonBuilder Null()
    {
        BeforeValue();
        _text.Append("null");
        return this;
    }

    /// <summary>Writes text that is already JSON, as it is.</summary>
    /// <param name="json">One JSON value.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Raw(string json)
    {
        ArgumentNullException.ThrowIfNull(json);
        BeforeValue();
        _text.Append(json);
        return this;
    }

    /// <summary>Writes any JSON value, with the keys of its objects sorted by their UTF-8 bytes.</summary>
    /// <param name="value">The value; <c>null</c> is written as <c>null</c>.</param>
    /// <returns>This builder.</returns>
    public JsonBuilder Any(JsonAny? value)
    {
        switch (value)
        {
            case null or JsonNull:
                return Null();
            case JsonBool boolean:
                return Boolean(boolean.Value);
            case JsonText text:
                return String(text.Value);
            case JsonNumber { Integer: { } integer }:
                return Integer(integer);
            case JsonNumber number:
                return Number(number.Real);
            case JsonList list:
                BeginArray();
                foreach (var item in list)
                {
                    Any(item);
                }

                return EndArray();
            case JsonMap map:
                BeginObject();
                var keys = map.Keys.ToList();
                keys.Sort(GoJsonTextAlgorithm.CompareUtf8);
                foreach (var key in keys)
                {
                    Property(key);
                    Any(map[key]);
                }

                return EndObject();
            default:
                throw new JsonEncodeError($"a {value.GetType().Name} cannot be written as JSON");
        }
    }

    /// <summary>The JSON written so far.</summary>
    /// <returns>The text.</returns>
    public override string ToString() => _text.ToString();

    private void BeforeValue()
    {
        if (_containers.Count == 0)
        {
            return;
        }

        var state = _containers[^1];
        if (!state.IsObject)
        {
            if (state.Count > 0)
            {
                _text.Append(',');
            }

            state.Count++;
        }
    }

    private sealed class ContainerState(bool IsObject)
    {
        public bool IsObject { get; } = IsObject;

        public int Count { get; set; }
    }
}
