using System.Text;
using MonoSmith.Plugin.Sdk.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>
/// A plugin served in this process over pipes, spoken to with raw frames. The test sends bodies exactly as it writes
/// them and reads the bodies the plugin wrote, so it can check the bytes. Nothing here sleeps: <see cref="Recv"/>
/// blocks until a frame arrives.
/// </summary>
internal sealed class Conversation : IDisposable
{
    /// <summary>How long a test waits for a frame before it fails. Never a pause the test depends on.</summary>
    public static readonly TimeSpan Wait = TimeSpan.FromSeconds(30);

    private readonly InProcessPipeStore _toPlugin = new();
    private readonly InProcessPipeStore _fromPlugin = new();
    private readonly Queue<string> _pending = new();
    private readonly FrameReader _reader = new();
    private readonly byte[] _buffer = new byte[64 * 1024];
    private readonly Task _serving;
    private Exception? _error;

    public Conversation(IPlugin plugin)
    {
        _serving = Task.Run(async () =>
        {
            try
            {
                await PluginServer.ServeAsync(plugin, new ServeStreams(_toPlugin.Reader, _fromPlugin.Writer, Logs.Add));
            }
            catch (Exception error)
            {
                _error = error;
            }
            finally
            {
                await _fromPlugin.Writer.DisposeAsync();
            }
        });
    }

    public List<string> Logs { get; } = [];

    /// <summary>Writes a JSON-RPC message as a frame.</summary>
    public void Send(string message) => SendRaw("{\"jsonrpc\":\"2.0\"," + message[1..]);

    /// <summary>Writes a body as a frame, as it is.</summary>
    public void SendRaw(string body) => SendBytes(Sdk.Framing.Frame(body));

    /// <summary>Writes bytes as they are, frame or not.</summary>
    public void SendBytes(byte[] data) => _toPlugin.Writer.Write(data, 0, data.Length);

    /// <summary>Writes a request.</summary>
    public void Request(long id, string method, string? paramsJson = null) =>
        Send($"{{\"id\":{id},\"method\":\"{method}\"" + (paramsJson is null ? string.Empty : $",\"params\":{paramsJson}") + "}");

    /// <summary>Writes a notification.</summary>
    public void Notify(string method, string paramsJson) => Send($"{{\"method\":\"{method}\",\"params\":{paramsJson}}}");

    /// <summary>The body of the next frame the plugin wrote; fails the test if none comes.</summary>
    public string Recv()
    {
        while (_pending.Count == 0)
        {
            var read = _fromPlugin.Reader.ReadAsync(_buffer).AsTask();
            Assert.True(read.Wait(Wait), "no frame came from the plugin");
            Assert.True(read.Result > 0, "the plugin closed its output before the expected frame");
            foreach (var body in _reader.Push(_buffer.AsSpan(0, read.Result)))
            {
                _pending.Enqueue(body);
            }
        }

        return _pending.Dequeue();
    }

    /// <summary>The next frame, parsed.</summary>
    public JsonMap RecvJson() => Assert.IsType<JsonMap>(JsonAny.Parse(Recv()));

    /// <summary>Every remaining frame, until the plugin closes its output.</summary>
    public List<string> RecvAll()
    {
        var bodies = new List<string>(_pending);
        _pending.Clear();
        while (true)
        {
            var read = _fromPlugin.Reader.ReadAsync(_buffer).AsTask();
            Assert.True(read.Wait(Wait), "the plugin did not close its output");
            if (read.Result == 0)
            {
                return bodies;
            }

            bodies.AddRange(_reader.Push(_buffer.AsSpan(0, read.Result)));
        }
    }

    /// <summary>Ends the plugin's input, as a host that is gone does.</summary>
    public void CloseInput() => _toPlugin.Writer.Dispose();

    /// <summary>Waits for the serve loop to end; rethrows what it threw.</summary>
    public void Finish()
    {
        Assert.True(_serving.Wait(Wait), "the serve loop did not end");
        if (_error is not null)
        {
            throw new InvalidOperationException("the serve loop failed", _error);
        }
    }

    /// <summary>Sends <c>initialize</c> with id 1 and returns the result.</summary>
    public JsonMap Handshake(string versions = "[1]")
    {
        Request(1, "initialize", $"{{\"protocolVersions\":{versions},\"host\":{{\"name\":\"smith\",\"version\":\"1\"}},\"workspaceRoot\":\"/w\"}}");
        var answer = RecvJson();
        Assert.True(answer.ContainsKey("result"), answer.Encode());
        return Assert.IsType<JsonMap>(answer["result"]);
    }

    public void Dispose()
    {
        CloseInput();
        Finish();
        _fromPlugin.Reader.Dispose();
    }

    /// <summary>The UTF-8 text of bytes, for a failure message.</summary>
    public static string Text(byte[] bytes) => Encoding.UTF8.GetString(bytes);
}
