namespace MonoSmith.Plugin.Sdk.Tests.Json;

public sealed class JsonTests
{
    [Theory]
    [InlineData("{}", "{}")]
    [InlineData("[ 1 , 2 ,3 ]", "[1,2,3]")]
    [InlineData("{\"b\":1,\"a\":[true,null,\"x\"]}", "{\"a\":[true,null,\"x\"],\"b\":1}")]
    [InlineData("  \"text\"  ", "\"text\"")]
    public void Encodes_compactly_with_the_keys_sorted_as_Go_writes_a_map(string text, string expected) =>
        Assert.Equal(expected, JsonAny.Parse(text).Encode());

    [Fact]
    public void A_key_that_repeats_keeps_its_first_place_and_its_last_value()
    {
        var map = Assert.IsType<JsonMap>(JsonAny.Parse("{\"a\":1,\"b\":2,\"a\":3}"));
        Assert.Equal(["a", "b"], map.Keys);
        Assert.Equal(new JsonNumber(3L), map["a"]);
    }

    [Fact]
    public void Sorts_keys_by_their_UTF8_bytes_not_their_UTF16_units()
    {
        // U+FF5E is below U+1F600 as a code point, and above it as UTF-16 units (a surrogate pair starts at D83D).
        var map = new JsonMap { [char.ConvertFromUtf32(0x1F600)] = 1L, [char.ConvertFromUtf32(0xFF5E)] = 2L };
        Assert.Equal("{\"" + char.ConvertFromUtf32(0xFF5E) + "\":2,\"" + char.ConvertFromUtf32(0x1F600) + "\":1}", map.Encode());
    }

    [Fact]
    public void Escapes_the_characters_Go_escapes()
    {
        var text = "<a&b>" + char.ConvertFromUtf32(0x2028) + char.ConvertFromUtf32(0x2029) + "\u0001\"\\\n\t\b\f" + "é";
        Assert.Equal("\"@Uu003ca@Uu0026b@Uu003e@Uu2028@Uu2029@Uu0001\\\"\\\\\\n\\t\\b\\fé\"".Replace("@U", "\\"), new JsonText(text).Encode());
    }

    [Fact]
    public void Reads_escapes_and_writes_a_lone_surrogate_as_the_replacement_character()
    {
        var parsed = Assert.IsType<JsonText>(JsonAny.Parse("\"@Uu00e9@Uud83d@Uude00 @Uud800\"".Replace("@U", "\\")));
        Assert.Equal("é" + char.ConvertFromUtf32(0x1F600) + " " + char.ConvertFromUtf32(0xFFFD), parsed.Value);
    }

    [Theory]
    [InlineData("1", "1")]
    [InlineData("1.0", "1")]
    [InlineData("-0", "0")]
    [InlineData("1.5", "1.5")]
    [InlineData("1e21", "1e+21")]
    [InlineData("100000000000000000000", "100000000000000000000")]
    [InlineData("0.000001", "0.000001")]
    [InlineData("1e-7", "1e-7")]
    [InlineData("12345678901234567890", "12345678901234567000")]
    [InlineData("9223372036854775807", "9223372036854775807")]
    public void Writes_numbers_the_way_Go_writes_them(string text, string expected) =>
        Assert.Equal(expected, JsonAny.Parse(text).Encode());

    [Fact]
    public void Keeps_an_integer_exactly_only_when_it_is_written_as_one()
    {
        Assert.Equal(9223372036854775807L, Assert.IsType<JsonNumber>(JsonAny.Parse("9223372036854775807")).Integer);
        Assert.Null(Assert.IsType<JsonNumber>(JsonAny.Parse("1.0")).Integer);
        Assert.Null(Assert.IsType<JsonNumber>(JsonAny.Parse("9223372036854775808")).Integer);
    }

    [Theory]
    [InlineData("")]
    [InlineData("{")]
    [InlineData("[1,]")]
    [InlineData("01")]
    [InlineData("NaN")]
    [InlineData("{\"a\":1} x")]
    [InlineData("\"unterminated")]
    [InlineData("'single'")]
    public void Rejects_text_that_is_not_one_JSON_value(string text) =>
        Assert.Throws<JsonParseError>(() => JsonAny.Parse(text));

    [Fact]
    public void Refuses_to_write_a_number_that_is_not_finite()
    {
        Assert.Throws<ArgumentOutOfRangeException>(() => new JsonNumber(double.NaN));
        Assert.Throws<JsonEncodeError>(() => new JsonBuilder().Number(double.PositiveInfinity));
    }

    [Fact]
    public void Describes_each_kind_of_value_for_an_error_message()
    {
        Assert.Equal("an object", new JsonMap().Describe());
        Assert.Equal("an array", new JsonList().Describe());
        Assert.Equal("a string", new JsonText("x").Describe());
        Assert.Equal("a number", new JsonNumber(1L).Describe());
        Assert.Equal("a boolean", new JsonBool(true).Describe());
        Assert.Equal("null", JsonAny.Describe(null));
    }

    [Fact]
    public void Values_compare_by_content()
    {
        Assert.Equal(JsonAny.Parse("{\"a\":[1,{\"b\":null}]}"), JsonAny.Parse("{ \"a\": [1, {\"b\": null}] }"));
        Assert.NotEqual(JsonAny.Parse("[1,2]"), JsonAny.Parse("[2,1]"));
    }

    [Fact]
    public void Refuses_a_text_nested_too_deeply_instead_of_overflowing_the_stack()
    {
        var deep = new string('[', 5000) + new string(']', 5000);
        Assert.Throws<JsonParseError>(() => JsonAny.Parse(deep));
    }
}
