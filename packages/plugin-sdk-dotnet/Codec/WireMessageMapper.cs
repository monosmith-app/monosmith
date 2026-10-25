namespace MonoSmith.Plugin.Sdk;

/// <summary>The JSON-RPC 2.0 envelopes: written the way the Go SDK writes them, and read.</summary>
internal static class WireMessageMapper
{
    /// <summary>A request as the Go SDK writes it: <c>jsonrpc</c>, <c>id</c>, <c>method</c>, <c>params</c> (left out when there are none).</summary>
    public static string WriteRequest(long id, string method, string? paramsJson)
    {
        var writer = new JsonBuilder().BeginObject();
        writer.Property("jsonrpc").String("2.0");
        writer.Property("id").Integer(id);
        writer.Property("method").String(method);
        if (paramsJson is not null)
        {
            writer.Property("params").Raw(paramsJson);
        }

        return writer.EndObject().ToString();
    }

    /// <summary>A notification as the Go SDK writes it: a request with no <c>id</c>, which gets no response.</summary>
    public static string WriteNotification(string method, string paramsJson)
    {
        var writer = new JsonBuilder().BeginObject();
        writer.Property("jsonrpc").String("2.0");
        writer.Property("method").String(method);
        writer.Property("params").Raw(paramsJson);
        return writer.EndObject().ToString();
    }

    /// <summary>A successful response; the id is <c>null</c> only when the request could not be read.</summary>
    public static string WriteResult(long? id, string resultJson)
    {
        var writer = new JsonBuilder().BeginObject();
        writer.Property("jsonrpc").String("2.0");
        WriteId(writer, id);
        writer.Property("result").Raw(resultJson);
        return writer.EndObject().ToString();
    }

    /// <summary>An error response.</summary>
    public static string WriteFailure(long? id, long code, string message)
    {
        var writer = new JsonBuilder().BeginObject();
        writer.Property("jsonrpc").String("2.0");
        WriteId(writer, id);
        writer.Property("error").BeginObject();
        writer.Property("code").Integer(code);
        writer.Property("message").String(message);
        writer.EndObject();
        return writer.EndObject().ToString();
    }

    /// <summary>Reads a request frame. A <c>null</c> or missing id makes it a notification.</summary>
    public static ReceivedRequest ParseRequest(string body)
    {
        var message = ParseObject(body);
        var id = ReadId(message);
        var method = message.TryGetValue("method", out var value) ? value : JsonNull.Instance;
        if (method is not (JsonNull or JsonText))
        {
            throw new MessageParseError("method must be a string");
        }

        message.TryGetValue("params", out var parameters);
        return new ReceivedRequest(id, method is JsonText text ? text.Value : string.Empty, parameters is JsonNull ? null : parameters);
    }

    /// <summary>Reads a response frame.</summary>
    public static ReceivedResponse ParseResponse(string body)
    {
        var message = ParseObject(body);
        var id = ReadId(message);
        if (!message.TryGetValue("error", out var error) || error is JsonNull)
        {
            message.TryGetValue("result", out var result);
            return new ReceivedResponse(id, result is JsonNull ? null : result, null);
        }

        if (error is not JsonMap fields)
        {
            throw new MessageParseError("error must be an object");
        }

        fields.TryGetValue("code", out var code);
        fields.TryGetValue("message", out var text);
        if (code is not JsonNumber { Integer: { } numeric } || text is not JsonText description)
        {
            throw new MessageParseError("error needs a numeric code and a message");
        }

        return new ReceivedResponse(id, null, new ReceivedError(numeric, description.Value));
    }

    private static void WriteId(JsonBuilder writer, long? id)
    {
        writer.Property("id");
        if (id is { } number)
        {
            writer.Integer(number);
        }
        else
        {
            writer.Null();
        }
    }

    private static JsonMap ParseObject(string body)
    {
        JsonAny value;
        try
        {
            value = JsonAny.Parse(body);
        }
        catch (JsonParseError error)
        {
            throw new MessageParseError(error.Message, error);
        }

        return value as JsonMap ?? throw new MessageParseError("a message is a JSON object");
    }

    private static long? ReadId(JsonMap message)
    {
        if (!message.TryGetValue("id", out var value))
        {
            return null;
        }

        return value switch
        {
            JsonNull => null,
            JsonNumber { Integer: { } integer } => integer,
            _ => throw new MessageParseError("id must be an integer"),
        };
    }
}
