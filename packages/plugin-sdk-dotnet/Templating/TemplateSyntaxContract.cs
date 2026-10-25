namespace MonoSmith.Plugin.Sdk;

/// <summary>One operand of a command: a field of the data, a literal, a parenthesised pipeline or a function name.</summary>
internal abstract record Operand;

/// <summary>A field of the data: <c>.A.B</c>, or <c>$.A</c> from the root.</summary>
internal sealed record FieldOperand(bool Root, IReadOnlyList<string> Path) : Operand;

/// <summary>A string, number or boolean literal.</summary>
internal sealed record LiteralOperand(JsonAny Value) : Operand;

/// <summary>The literal <c>nil</c>.</summary>
internal sealed record NilOperand : Operand;

/// <summary>A function name.</summary>
internal sealed record NameOperand(string Name) : Operand;

/// <summary>A parenthesised pipeline.</summary>
internal sealed record PipelineOperand(IReadOnlyList<IReadOnlyList<Operand>> Pipeline) : Operand;

/// <summary>A node of a parsed template.</summary>
internal abstract record TemplateNode;

/// <summary>Literal text.</summary>
internal sealed record TextNode(string Text) : TemplateNode;

/// <summary><c>[[ pipeline ]]</c>: a value printed.</summary>
internal sealed record ActionNode(IReadOnlyList<IReadOnlyList<Operand>> Pipeline) : TemplateNode;

/// <summary>What a block does with its pipeline.</summary>
internal enum BlockKind
{
    /// <summary>Renders its body when the value is true.</summary>
    If,

    /// <summary>Renders its body with the value as the dot when the value is true.</summary>
    With,

    /// <summary>Renders its body once for each element.</summary>
    Range,
}

/// <summary><c>if</c>, <c>with</c> or <c>range</c>, with the nodes of its body and of its <c>else</c>.</summary>
internal sealed record BlockNode(
    BlockKind Kind,
    IReadOnlyList<IReadOnlyList<Operand>> Pipeline,
    IReadOnlyList<TemplateNode> Body,
    IReadOnlyList<TemplateNode> Otherwise) : TemplateNode;
