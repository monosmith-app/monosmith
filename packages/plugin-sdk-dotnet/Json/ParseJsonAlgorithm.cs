using System.Globalization;
using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// A JSON parser with the semantics of Go's decoder where they differ from <c>System.Text.Json</c>: a key that
/// repeats keeps its first place and its last value, a lone surrogate escape becomes U+FFFD, and a number
/// is an integer (kept exactly) only when it is written as one.
/// </summary>
internal static class ParseJsonAlgorithm
{
    private const int MaxDepth = 1000;

    public static JsonAny Parse(string text)
    {
        ArgumentNullException.ThrowIfNull(text);
        var parser = new Parser(text);
        parser.SkipWhitespace();
        var value = parser.ReadValue(0);
        parser.SkipWhitespace();
        if (!parser.AtEnd)
        {
            throw parser.Fail("unexpected data after the JSON value");
        }

        return value;
    }

    private sealed class Parser(string text)
    {
        private int _position;

        public bool AtEnd => _position >= text.Length;

        public JsonParseError Fail(string message) =>
            new($"{message} at position {_position.ToString(CultureInfo.InvariantCulture)}");

        public void SkipWhitespace()
        {
            while (_position < text.Length && text[_position] is ' ' or '\t' or '\n' or '\r')
            {
                _position++;
            }
        }

        public JsonAny ReadValue(int depth)
        {
            if (depth > MaxDepth)
            {
                throw Fail("the JSON is nested too deeply");
            }

            if (AtEnd)
            {
                throw Fail("unexpected end of the JSON text");
            }

            switch (text[_position])
            {
                case '{':
                    return ReadObject(depth);
                case '[':
                    return ReadArray(depth);
                case '"':
                    return new JsonText(ReadString());
                case 't':
                    ReadWord("true");
                    return new JsonBool(true);
                case 'f':
                    ReadWord("false");
                    return new JsonBool(false);
                case 'n':
                    ReadWord("null");
                    return JsonNull.Instance;
                default:
                    return ReadNumber();
            }
        }

        private JsonMap ReadObject(int depth)
        {
            var map = new JsonMap();
            _position++;
            SkipWhitespace();
            if (Peek() == '}')
            {
                _position++;
                return map;
            }

            while (true)
            {
                SkipWhitespace();
                if (Peek() != '"')
                {
                    throw Fail("expected a string key");
                }

                var key = ReadString();
                SkipWhitespace();
                if (Peek() != ':')
                {
                    throw Fail("expected ':' after a key");
                }

                _position++;
                SkipWhitespace();
                map[key] = ReadValue(depth + 1);
                SkipWhitespace();
                var following = Peek();
                _position++;
                if (following == '}')
                {
                    return map;
                }

                if (following != ',')
                {
                    _position--;
                    throw Fail("expected ',' or '}' in an object");
                }
            }
        }

        private JsonList ReadArray(int depth)
        {
            var list = new JsonList();
            _position++;
            SkipWhitespace();
            if (Peek() == ']')
            {
                _position++;
                return list;
            }

            while (true)
            {
                SkipWhitespace();
                list.Add(ReadValue(depth + 1));
                SkipWhitespace();
                var following = Peek();
                _position++;
                if (following == ']')
                {
                    return list;
                }

                if (following != ',')
                {
                    _position--;
                    throw Fail("expected ',' or ']' in an array");
                }
            }
        }

        private string ReadString()
        {
            _position++;
            var builder = new StringBuilder();
            while (true)
            {
                if (AtEnd)
                {
                    throw Fail("unterminated string");
                }

                var c = text[_position++];
                if (c == '"')
                {
                    return builder.ToString();
                }

                if (c < ' ')
                {
                    _position--;
                    throw Fail("a control character in a string");
                }

                if (c != '\\')
                {
                    builder.Append(c);
                    continue;
                }

                if (AtEnd)
                {
                    throw Fail("unterminated string");
                }

                var escape = text[_position++];
                switch (escape)
                {
                    case '"' or '\\' or '/':
                        builder.Append(escape);
                        break;
                    case 'b':
                        builder.Append('\b');
                        break;
                    case 'f':
                        builder.Append('\f');
                        break;
                    case 'n':
                        builder.Append('\n');
                        break;
                    case 'r':
                        builder.Append('\r');
                        break;
                    case 't':
                        builder.Append('\t');
                        break;
                    case 'u':
                        builder.Append(ReadUnicodeEscape());
                        break;
                    default:
                        _position--;
                        throw Fail("an invalid escape in a string");
                }
            }
        }

        private char[] ReadUnicodeEscape()
        {
            var first = ReadHex4();
            if (first is >= 0xD800 and <= 0xDBFF)
            {
                if (_position + 1 < text.Length && text[_position] == '\\' && text[_position + 1] == 'u')
                {
                    var saved = _position;
                    _position += 2;
                    var second = ReadHex4();
                    if (second is >= 0xDC00 and <= 0xDFFF)
                    {
                        return [(char)first, (char)second];
                    }

                    _position = saved;
                }

                return ['�'];
            }

            return first is >= 0xDC00 and <= 0xDFFF ? ['�'] : [(char)first];
        }

        private int ReadHex4()
        {
            if (_position + 4 > text.Length || !int.TryParse(text.AsSpan(_position, 4), NumberStyles.AllowHexSpecifier, CultureInfo.InvariantCulture, out var value)
                || text.AsSpan(_position, 4).IndexOfAny('+', '-', ' ') >= 0)
            {
                throw Fail("an invalid \\u escape in a string");
            }

            _position += 4;
            return value;
        }

        private void ReadWord(string word)
        {
            if (string.CompareOrdinal(text, _position, word, 0, word.Length) != 0)
            {
                throw Fail($"invalid character, expected '{word}'");
            }

            _position += word.Length;
        }

        private JsonNumber ReadNumber()
        {
            var start = _position;
            if (Peek() == '-')
            {
                _position++;
            }

            if (Peek() == '0')
            {
                _position++;
            }
            else if (Peek() is >= '1' and <= '9')
            {
                SkipDigits();
            }
            else
            {
                _position = start;
                throw Fail("invalid character, expected a JSON value");
            }

            var integer = true;
            if (Peek() == '.')
            {
                integer = false;
                _position++;
                RequireDigits();
            }

            if (Peek() is 'e' or 'E')
            {
                integer = false;
                _position++;
                if (Peek() is '+' or '-')
                {
                    _position++;
                }

                RequireDigits();
            }

            var literal = text.AsSpan(start, _position - start);
            if (integer && long.TryParse(literal, NumberStyles.AllowLeadingSign, CultureInfo.InvariantCulture, out var whole))
            {
                return new JsonNumber(whole);
            }

            var real = double.Parse(literal, NumberStyles.Float, CultureInfo.InvariantCulture);
            if (!double.IsFinite(real))
            {
                _position = start;
                throw Fail("a number out of range");
            }

            return new JsonNumber(real);
        }

        private void SkipDigits()
        {
            while (Peek() is >= '0' and <= '9')
            {
                _position++;
            }
        }

        private void RequireDigits()
        {
            if (Peek() is not (>= '0' and <= '9'))
            {
                throw Fail("invalid number, expected a digit");
            }

            SkipDigits();
        }

        private char Peek() => _position < text.Length ? text[_position] : '\0';
    }
}
