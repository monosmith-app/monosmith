using System.Globalization;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// A JSON number: an integer that fits 64 bits is kept exactly, anything else is a double, as Go reads
/// a number into <c>any</c>.
/// </summary>
public sealed record JsonNumber : JsonAny
{
    /// <summary>An integer.</summary>
    /// <param name="value">The integer.</param>
    public JsonNumber(long value)
    {
        Integer = value;
        Real = value;
    }

    /// <summary>A floating-point number.</summary>
    /// <param name="value">The number; it must be finite.</param>
    public JsonNumber(double value)
    {
        if (!double.IsFinite(value))
        {
            throw new ArgumentOutOfRangeException(nameof(value), value, "a JSON number is finite");
        }

        Real = value;
    }

    /// <summary>The value, when it was written as an integer that fits 64 bits.</summary>
    public long? Integer { get; }

    /// <summary>The value as a double.</summary>
    public double Real { get; }

    /// <inheritdoc />
    public override string Describe() => "a number";

    /// <inheritdoc />
    public override string ToString() =>
        Integer is { } integer ? integer.ToString(CultureInfo.InvariantCulture) : GoJsonTextAlgorithm.FormatNumber(Real);
}
