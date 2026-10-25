namespace MonoSmith.Plugin.Sdk.Protocol;

/// <summary>Writing lists and maps for the writers the generator writes: keys sorted by their UTF-8 bytes.</summary>
internal static class MessageWritingAlgorithm
{
    /// <summary>Writes an array. A <c>null</c> list is an empty one.</summary>
    public static void List<T>(JsonBuilder writer, IReadOnlyList<T>? items, Action<JsonBuilder, T> write)
    {
        writer.BeginArray();
        if (items is not null)
        {
            foreach (var item in items)
            {
                write(writer, item);
            }
        }

        writer.EndArray();
    }

    /// <summary>Writes an object with its keys sorted, as Go writes a map. A <c>null</c> map is an empty one.</summary>
    public static void Map<T>(JsonBuilder writer, IReadOnlyDictionary<string, T>? members, Action<JsonBuilder, T> write)
    {
        writer.BeginObject();
        if (members is not null)
        {
            var keys = members.Keys.ToList();
            keys.Sort(GoJsonTextAlgorithm.CompareUtf8);
            foreach (var key in keys)
            {
                writer.Property(key);
                write(writer, members[key]);
            }
        }

        writer.EndObject();
    }
}
