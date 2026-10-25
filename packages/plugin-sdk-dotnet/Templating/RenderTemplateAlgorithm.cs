using System.Text;
using static MonoSmith.Plugin.Sdk.TemplateFunctionsAlgorithm;

namespace MonoSmith.Plugin.Sdk;

/// <summary><c>[[ ]]</c> templates, written for new projects and never over existing files.</summary>
public static partial class Templates
{
    /// <summary>Stands for "nothing piped". Compared by reference.</summary>
    private static readonly JsonText NothingPiped = new("nothing piped");

    /// <summary>
    /// Renders a <c>[[ ]]</c> template with the data, the way the Go SDK's <c>templating</c> does with
    /// <c>text/template</c> and <c>missingkey=error</c>: a field the data does not have is an error, not an empty string.
    /// </summary>
    /// <remarks>
    /// The language is the subset of Go's <c>text/template</c> that project templates use: <c>[[.Field]]</c> and
    /// <c>[[.A.B]]</c>, <c>[[if]]</c>, <c>[[else if]]</c>, <c>[[else]]</c>, <c>[[with]]</c>, <c>[[range]]</c> and
    /// <c>[[end]]</c>, the functions <c>eq ne lt le gt ge not and or len index print printf</c>, string, number and
    /// boolean literals, parentheses, pipes, comments, and the <c>-</c> trim markers (<c>[[- x -]]</c>). Variables,
    /// <c>define</c>, <c>template</c> and <c>block</c> are not supported and are reported as such.
    /// </remarks>
    /// <param name="text">The template.</param>
    /// <param name="data">The data: a <see cref="JsonMap"/> of the fields the template names.</param>
    /// <param name="name">What error messages call the template.</param>
    /// <returns>The rendered text.</returns>
    /// <exception cref="TemplateError">The template cannot be parsed or rendered.</exception>
    public static string Render(string text, JsonAny? data = null, string name = "template")
    {
        ArgumentNullException.ThrowIfNull(text);
        TemplateError Fail(string message) => new(name, message);
        return RenderNodes(ParseTemplateAlgorithm.Parse(text, name), new Scope(data, data), Fail);
    }

    private static string RenderNodes(IReadOnlyList<TemplateNode> nodes, Scope scope, Func<string, TemplateError> fail)
    {
        var output = new StringBuilder();
        foreach (var node in nodes)
        {
            output.Append(RenderNode(node, scope, fail));
        }

        return output.ToString();
    }

    private static string RenderNode(TemplateNode node, Scope scope, Func<string, TemplateError> fail)
    {
        if (node is TextNode text)
        {
            return text.Text;
        }

        if (node is ActionNode action)
        {
            return FormatValue(EvaluatePipeline(action.Pipeline, scope, fail));
        }

        var block = (BlockNode)node;
        var value = EvaluatePipeline(block.Pipeline, scope, fail);
        switch (block.Kind)
        {
            case BlockKind.If:
                return RenderNodes(IsTruthy(value) ? block.Body : block.Otherwise, scope, fail);
            case BlockKind.With:
                return IsTruthy(value)
                    ? RenderNodes(block.Body, new Scope(value, scope.Root), fail)
                    : RenderNodes(block.Otherwise, scope, fail);
            default:
                var items = RangeItems(value, fail);
                if (items.Count == 0)
                {
                    return RenderNodes(block.Otherwise, scope, fail);
                }

                var output = new StringBuilder();
                foreach (var item in items)
                {
                    output.Append(RenderNodes(block.Body, new Scope(item, scope.Root), fail));
                }

                return output.ToString();
        }
    }

    /// <summary>What <c>range</c> loops over: the elements of a list, or the values of a map in key order.</summary>
    private static List<JsonAny?> RangeItems(JsonAny? value, Func<string, TemplateError> fail) => value switch
    {
        JsonList list => [.. list],
        null or JsonNull => [],
        JsonMap map => [.. SortedKeys(map).Select(key => (JsonAny?)map[key])],
        _ => throw fail($"range can't iterate over {FormatValue(value)}"),
    };

    private static JsonAny? EvaluatePipeline(IReadOnlyList<IReadOnlyList<Operand>> pipeline, Scope scope, Func<string, TemplateError> fail)
    {
        JsonAny? piped = NothingPiped;
        foreach (var command in pipeline)
        {
            piped = EvaluateCommand(command, scope, piped, fail);
        }

        return piped;
    }

    private static JsonAny? EvaluateCommand(IReadOnlyList<Operand> command, Scope scope, JsonAny? piped, Func<string, TemplateError> fail)
    {
        var head = command[0];
        var rest = command.Skip(1).ToList();
        if (head is NameOperand name)
        {
            return CallFunction(name.Name, rest, scope, piped, fail);
        }

        if (rest.Count > 0)
        {
            throw fail("can't give an argument to a value that is not a function");
        }

        if (ReferenceEquals(piped, NothingPiped))
        {
            return EvaluateOperand(head, scope, fail);
        }

        throw fail("can't pipe into a value that is not a function");
    }

    private static JsonAny? CallFunction(string name, List<Operand> operands, Scope scope, JsonAny? piped, Func<string, TemplateError> fail)
    {
        if (!Functions.TryGetValue(name, out var function))
        {
            throw fail($"function \"{name}\" not defined");
        }

        var arguments = operands.Select(operand => EvaluateOperand(operand, scope, fail)).ToList();
        if (!ReferenceEquals(piped, NothingPiped))
        {
            arguments.Add(piped);
        }

        try
        {
            return function(arguments);
        }
        catch (TemplateFunctionError error)
        {
            throw fail($"error calling {name}: {error.Message}");
        }
        catch (ArgumentOutOfRangeException)
        {
            throw fail($"error calling {name}: wrong number of args");
        }
        catch (InvalidOperationException)
        {
            throw fail($"error calling {name}: wrong number of args");
        }
    }

    private static JsonAny? EvaluateOperand(Operand operand, Scope scope, Func<string, TemplateError> fail)
    {
        switch (operand)
        {
            case NilOperand:
                return null;
            case LiteralOperand literal:
                return literal.Value;
            case PipelineOperand nested:
                return EvaluatePipeline(nested.Pipeline, scope, fail);
            case NameOperand name:
                return EvaluateCommand([name], scope, NothingPiped, fail);
            default:
                var field = (FieldOperand)operand;
                var value = field.Root ? scope.Root : scope.Dot;
                foreach (var key in field.Path)
                {
                    if (value is not JsonMap map)
                    {
                        throw fail($"can't evaluate field {key} in {(IsNil(value) ? "nil" : TypeName(value!))}");
                    }

                    if (!map.TryGetValue(key, out var member))
                    {
                        throw fail($"map has no entry for key \"{key}\"");
                    }

                    value = member;
                }

                return value;
        }
    }

    private static string TypeName(JsonAny value) => value switch
    {
        JsonBool => "boolean",
        JsonText => "string",
        JsonNumber => "number",
        _ => "object",
    };

    private sealed record Scope(JsonAny? Dot, JsonAny? Root);
}
