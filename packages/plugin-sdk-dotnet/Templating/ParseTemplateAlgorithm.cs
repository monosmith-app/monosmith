namespace MonoSmith.Plugin.Sdk;

/// <summary>Parsing a template with <c>[[ ]]</c> delimiters.</summary>
/// <remarks>
/// The language is the subset of Go's <c>text/template</c> that project templates use: <c>[[.Field]]</c> and
/// <c>[[.A.B]]</c>, <c>[[if]]</c>, <c>[[else if]]</c>, <c>[[else]]</c>, <c>[[with]]</c>, <c>[[range]]</c> and
/// <c>[[end]]</c>, the functions <c>eq ne lt le gt ge not and or len index print printf</c>, string, number and
/// boolean literals, parentheses, pipes, comments, and the <c>-</c> trim markers (<c>[[- x -]]</c>). Variables,
/// <c>define</c>, <c>template</c> and <c>block</c> are not supported and are reported as such.
/// </remarks>
internal static class ParseTemplateAlgorithm
{
    private const string Open = "[[";
    private const string Close = "]]";
    private static readonly char[] Space = [' ', '\t', '\r', '\n'];
    private static readonly HashSet<string> Unsupported = ["define", "template", "block", "break", "continue"];

    public static IReadOnlyList<TemplateNode> Parse(string text, string name)
    {
        TemplateError Fail(string message) => new(name, message);

        var root = new List<TemplateNode>();
        var stack = new List<Frame>();

        List<TemplateNode> Target() => stack.Count == 0 ? root : stack[^1].InElse ? stack[^1].Otherwise : stack[^1].Body;

        foreach (var (literal, action) in SplitPieces(text, Fail))
        {
            if (action is null)
            {
                if (literal.Length > 0)
                {
                    Target().Add(new TextNode(literal));
                }
            }
            else if (action.StartsWith("/*", StringComparison.Ordinal))
            {
                if (!action.EndsWith("*/", StringComparison.Ordinal))
                {
                    throw Fail("unclosed comment");
                }
            }
            else
            {
                ApplyAction(action, stack, Target, Fail);
            }
        }

        if (stack.Count > 0)
        {
            throw Fail($"unclosed {stack[^1].Kind.ToString().ToLowerInvariant()}: missing [[end]]");
        }

        return root;
    }

    /// <summary>The literal text and the actions of a template, in order, with the <c>-</c> trim markers applied.</summary>
    private static List<(string Literal, string? Action)> SplitPieces(string text, Func<string, TemplateError> fail)
    {
        var pieces = new List<(string, string?)>();
        var position = 0;
        var trimNext = false;
        while (position < text.Length)
        {
            var openAt = text.IndexOf(Open, position, StringComparison.Ordinal);
            var literal = text[position..(openAt == -1 ? text.Length : openAt)];
            if (trimNext)
            {
                literal = literal.TrimStart(Space);
            }

            trimNext = false;
            if (openAt == -1)
            {
                pieces.Add((literal, null));
                break;
            }

            var close = FindClose(text, openAt + Open.Length);
            if (close == -1)
            {
                throw fail($"unclosed action at offset {openAt}");
            }

            var action = text[(openAt + Open.Length)..close];
            if (action.Length >= 2 && action[0] == '-' && Array.IndexOf(Space, action[1]) >= 0)
            {
                literal = literal.TrimEnd(Space);
                action = action[1..];
            }

            if (action.Length >= 2 && action[^1] == '-' && Array.IndexOf(Space, action[^2]) >= 0)
            {
                trimNext = true;
                action = action[..^1];
            }

            pieces.Add((literal, null));
            pieces.Add((string.Empty, action.Trim(Space)));
            position = close + Close.Length;
        }

        return pieces;
    }

    /// <summary>The offset of the <c>]]</c> that ends the action, skipping string literals.</summary>
    private static int FindClose(string text, int start)
    {
        var i = start;
        while (i < text.Length)
        {
            var c = text[i];
            if (c == '"')
            {
                i++;
                while (i < text.Length && text[i] != '"')
                {
                    i += text[i] == '\\' ? 2 : 1;
                }
            }
            else if (c == '`')
            {
                i = text.IndexOf('`', i + 1);
                if (i == -1)
                {
                    return -1;
                }
            }
            else if (string.CompareOrdinal(text, i, Close, 0, Close.Length) == 0)
            {
                return i;
            }

            i++;
        }

        return -1;
    }

    private static void ApplyAction(string action, List<Frame> stack, Func<List<TemplateNode>> target, Func<string, TemplateError> fail)
    {
        var tokens = TokenizeActionAlgorithm.Tokenize(action, fail);
        var first = tokens.Count > 0 ? tokens[0] : null;
        var keyword = first is { Kind: TokenKind.Word, Value: JsonText word } ? word.Value : string.Empty;
        switch (keyword)
        {
            case "if" or "with" or "range":
                var kind = keyword == "if" ? BlockKind.If : keyword == "with" ? BlockKind.With : BlockKind.Range;
                stack.Add(new Frame(kind, ParsePipeline([.. tokens.Skip(1)], fail)));
                break;
            case "else":
                StartElse(tokens, stack, fail);
                break;
            case "end":
                if (tokens.Count > 1)
                {
                    throw fail("unexpected tokens after end");
                }

                CloseFrames(stack, target, fail);
                break;
            default:
                if (Unsupported.Contains(keyword))
                {
                    throw fail(
                        $"\"{keyword}\" is not supported by the SDK's templates (the supported subset: fields, if, else, with, range, end, " +
                        "and the functions eq ne lt le gt ge not and or len index print printf)");
                }

                if (tokens.Count == 0)
                {
                    throw fail("empty action");
                }

                target().Add(new ActionNode(ParsePipeline(tokens, fail)));
                break;
        }
    }

    private static void StartElse(IReadOnlyList<Token> tokens, List<Frame> stack, Func<string, TemplateError> fail)
    {
        if (stack.Count == 0)
        {
            throw fail("unexpected else");
        }

        var top = stack[^1];
        if (top.InElse)
        {
            throw fail("multiple else in one block");
        }

        top.InElse = true;
        var following = tokens.Count > 1 ? tokens[1] : null;
        if (following is { Kind: TokenKind.Word, Value: JsonText { Value: "if" } })
        {
            stack.Add(new Frame(BlockKind.If, ParsePipeline([.. tokens.Skip(2)], fail)) { Chained = true });
        }
        else if (tokens.Count > 1)
        {
            throw fail("unexpected tokens after else");
        }
    }

    /// <summary>Ends the innermost block, and the <c>else if</c> blocks chained to it.</summary>
    private static void CloseFrames(List<Frame> stack, Func<List<TemplateNode>> target, Func<string, TemplateError> fail)
    {
        while (true)
        {
            if (stack.Count == 0)
            {
                throw fail("unexpected end");
            }

            var frame = stack[^1];
            stack.RemoveAt(stack.Count - 1);
            target().Add(new BlockNode(frame.Kind, frame.Pipeline, frame.Body, frame.Otherwise));
            if (!frame.Chained)
            {
                return;
            }
        }
    }

    private static List<IReadOnlyList<Operand>> ParsePipeline(IReadOnlyList<Token> tokens, Func<string, TemplateError> fail)
    {
        var (pipeline, used) = PipelineFrom(tokens, 0, fail);
        if (used < tokens.Count)
        {
            throw fail("unexpected \")\" in action");
        }

        return pipeline;
    }

    private static (List<IReadOnlyList<Operand>> Pipeline, int Used) PipelineFrom(IReadOnlyList<Token> tokens, int start, Func<string, TemplateError> fail)
    {
        var pipeline = new List<IReadOnlyList<Operand>>();
        var command = new List<Operand>();
        var i = start;
        while (i < tokens.Count)
        {
            var token = tokens[i];
            if (token.Kind == TokenKind.Close)
            {
                break;
            }

            if (token.Kind == TokenKind.Pipe)
            {
                if (command.Count == 0)
                {
                    throw fail("missing value before |");
                }

                pipeline.Add(command);
                command = [];
            }
            else if (token.Kind == TokenKind.Open)
            {
                var (inner, following) = PipelineFrom(tokens, i + 1, fail);
                if (following >= tokens.Count || tokens[following].Kind != TokenKind.Close)
                {
                    throw fail("unclosed (");
                }

                command.Add(new PipelineOperand(inner));
                i = following;
            }
            else
            {
                command.Add(OperandOf(token));
            }

            i++;
        }

        if (command.Count == 0)
        {
            throw fail(pipeline.Count > 0 ? "missing command after |" : "missing value for the action");
        }

        pipeline.Add(command);
        return (pipeline, i);
    }

    private static Operand OperandOf(Token token)
    {
        switch (token.Kind)
        {
            case TokenKind.Field:
                return new FieldOperand(token.Root, token.Path ?? []);
            case TokenKind.String or TokenKind.Number:
                return new LiteralOperand(token.Value!);
            default:
                var word = ((JsonText)token.Value!).Value;
                return word switch
                {
                    "true" => new LiteralOperand(new JsonBool(true)),
                    "false" => new LiteralOperand(new JsonBool(false)),
                    "nil" => new NilOperand(),
                    _ => new NameOperand(word),
                };
        }
    }

    /// <summary>An <c>if</c>, <c>with</c> or <c>range</c> whose <c>[[end]]</c> has not come yet.</summary>
    private sealed class Frame(BlockKind kind, IReadOnlyList<IReadOnlyList<Operand>> pipeline)
    {
        public BlockKind Kind { get; } = kind;

        public IReadOnlyList<IReadOnlyList<Operand>> Pipeline { get; } = pipeline;

        public List<TemplateNode> Body { get; } = [];

        public List<TemplateNode> Otherwise { get; } = [];

        public bool InElse { get; set; }

        /// <summary>Created by <c>else if</c>: closed together with the frame that holds it.</summary>
        public bool Chained { get; init; }
    }
}
