namespace MonoSmith.Plugin.Sdk;

/// <summary>A JSON value as the document keeps it: an object edited member by member, or the text of anything else.</summary>
internal abstract record DocumentNode;

/// <summary>A JSON object edited member by member.</summary>
internal sealed record ObjectNode(IReadOnlyList<KeyValuePair<string, DocumentNode>> Members) : DocumentNode;

/// <summary>Anything else, as the text it was written in.</summary>
internal sealed record RawNode(string Text) : DocumentNode;

/// <summary>Reading the text of a JSON object into a tree that keeps what you do not touch.</summary>
internal static class ScanJsonObjectAlgorithm
{
    private const string Whitespace = " \t\n\r";

    /// <summary>
    /// Reads the text of a JSON object into a tree that keeps the order of its members and the exact text of every leaf
    /// and array (<c>1.0</c> stays <c>1.0</c>). A duplicate key keeps its first position and takes its last value, as
    /// the Go SDK's document does.
    /// </summary>
    public static ObjectNode Scan(string text)
    {
        var scanner = new Scanner(text);
        scanner.Skip();
        if (scanner.At() != '{')
        {
            throw new JsonDocumentError("not a JSON object");
        }

        var root = scanner.ReadObject();
        scanner.Skip();
        if (scanner.Position < text.Length)
        {
            throw new JsonDocumentError($"unexpected \"{text[scanner.Position]}\" after the object, at {scanner.Position}");
        }

        return root;
    }

    /// <summary>The value of a piece of JSON text, as the JSON parser reads it.</summary>
    public static JsonAny ParseText(string text)
    {
        try
        {
            return JsonAny.Parse(text);
        }
        catch (JsonParseError error)
        {
            throw new JsonDocumentError(error.Message, error);
        }
    }

    private sealed class Scanner(string text)
    {
        public int Position { get; private set; }

        public char At() => Position < text.Length ? text[Position] : '\0';

        public void Skip()
        {
            while (Position < text.Length && Whitespace.Contains(text[Position], StringComparison.Ordinal))
            {
                Position++;
            }
        }

        public ObjectNode ReadObject()
        {
            var members = new List<KeyValuePair<string, DocumentNode>>();
            var index = new Dictionary<string, int>(StringComparer.Ordinal);
            Position++; // {
            Skip();
            if (At() == '}')
            {
                Position++;
                return new ObjectNode(members);
            }

            while (true)
            {
                Skip();
                if (At() != '"')
                {
                    throw Fail("an object key");
                }

                var key = ((JsonText)ParseText(ReadString())).Value;
                Skip();
                if (At() != ':')
                {
                    throw Fail("\":\"");
                }

                Position++;
                var value = ReadValue();
                if (index.TryGetValue(key, out var at))
                {
                    members[at] = new KeyValuePair<string, DocumentNode>(key, value);
                }
                else
                {
                    index[key] = members.Count;
                    members.Add(new KeyValuePair<string, DocumentNode>(key, value));
                }

                Skip();
                var following = At();
                if (following == '}')
                {
                    Position++;
                    return new ObjectNode(members);
                }

                if (following != ',')
                {
                    throw Fail("\",\" or \"}\"");
                }

                Position++;
            }
        }

        private DocumentNode ReadValue()
        {
            Skip();
            var first = At();
            if (first == '{')
            {
                return ReadObject();
            }

            var start = Position;
            if (first == '[')
            {
                SkipArray();
            }
            else if (first == '"')
            {
                ReadString();
            }
            else
            {
                SkipLiteral();
            }

            var literal = text[start..Position];
            try
            {
                ParseText(literal);
            }
            catch (JsonDocumentError error)
            {
                var shown = literal.Length > 40 ? literal[..40] : literal;
                throw new JsonDocumentError($"invalid JSON value \"{shown}\" at {start}", error);
            }

            return new RawNode(literal);
        }

        /// <summary>Moves past an array, whose elements stay text.</summary>
        private void SkipArray()
        {
            var depth = 0;
            while (Position < text.Length)
            {
                var c = text[Position];
                if (c == '"')
                {
                    ReadString();
                    continue;
                }

                if (c is '[' or '{')
                {
                    depth++;
                }
                else if (c is ']' or '}')
                {
                    depth--;
                }

                Position++;
                if (depth == 0)
                {
                    return;
                }
            }

            throw new JsonDocumentError("unterminated array");
        }

        /// <summary>Moves past a string literal and returns its text, quotes included.</summary>
        private string ReadString()
        {
            var start = Position;
            Position++;
            while (Position < text.Length)
            {
                var c = text[Position];
                Position++;
                if (c == '\\')
                {
                    Position++;
                }
                else if (c == '"')
                {
                    return text[start..Position];
                }
            }

            throw new JsonDocumentError("unterminated string");
        }

        private void SkipLiteral()
        {
            var start = Position;
            while (Position < text.Length && !Whitespace.Contains(text[Position], StringComparison.Ordinal) && ",}]".IndexOf(text[Position], StringComparison.Ordinal) < 0)
            {
                Position++;
            }

            if (Position == start)
            {
                throw Fail("a value");
            }
        }

        private JsonDocumentError Fail(string expected)
        {
            var found = Position < text.Length ? $"\"{text[Position]}\"" : "the end of the text";
            return new JsonDocumentError($"expected {expected}, found {found} at {Position}");
        }
    }
}
