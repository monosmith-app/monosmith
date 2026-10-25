using System.Threading.Channels;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>
/// A one-way byte pipe inside the process: what one side writes the other reads, and closing the write side is the
/// end of the stream. It lets a test serve a plugin over the real framing with no operating system pipe.
/// </summary>
internal sealed class InProcessPipeStore
{
    private readonly Channel<byte[]> _chunks = Channel.CreateUnbounded<byte[]>(new UnboundedChannelOptions { SingleReader = true });
    private byte[]? _current;
    private int _offset;
    private volatile bool _readerClosed;

    public InProcessPipeStore()
    {
        Reader = new ReadEnd(this);
        Writer = new WriteEnd(this);
    }

    /// <summary>What is read.</summary>
    public Stream Reader { get; }

    /// <summary>What is written.</summary>
    public Stream Writer { get; }

    private async ValueTask<int> ReadAsync(Memory<byte> destination, CancellationToken cancellation)
    {
        while (true)
        {
            if (_current is null)
            {
                if (_readerClosed)
                {
                    return 0;
                }

                if (!_chunks.Reader.TryRead(out _current))
                {
                    if (!await _chunks.Reader.WaitToReadAsync(cancellation).ConfigureAwait(false))
                    {
                        return 0;
                    }

                    continue;
                }

                _offset = 0;
            }

            var count = Math.Min(destination.Length, _current.Length - _offset);
            _current.AsMemory(_offset, count).CopyTo(destination);
            _offset += count;
            if (_offset == _current.Length)
            {
                _current = null;
            }

            return count;
        }
    }

    private void Write(ReadOnlySpan<byte> data)
    {
        if (_readerClosed || !_chunks.Writer.TryWrite(data.ToArray()))
        {
            throw new IOException("the pipe is closed");
        }
    }

    private void CloseWriter() => _chunks.Writer.TryComplete();

    private void CloseReader()
    {
        _readerClosed = true;
        _chunks.Writer.TryComplete();
    }

    private sealed class ReadEnd(InProcessPipeStore pipe) : Stream
    {
        public override bool CanRead => true;

        public override bool CanSeek => false;

        public override bool CanWrite => false;

        public override long Length => throw new NotSupportedException();

        public override long Position
        {
            get => throw new NotSupportedException();
            set => throw new NotSupportedException();
        }

        public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken cancellationToken = default) =>
            pipe.ReadAsync(buffer, cancellationToken);

        public override int Read(byte[] buffer, int offset, int count) =>
            pipe.ReadAsync(buffer.AsMemory(offset, count), CancellationToken.None).AsTask().GetAwaiter().GetResult();

        public override void Flush()
        {
        }

        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();

        public override void SetLength(long value) => throw new NotSupportedException();

        public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();

        protected override void Dispose(bool disposing)
        {
            if (disposing)
            {
                pipe.CloseReader();
            }

            base.Dispose(disposing);
        }
    }

    private sealed class WriteEnd(InProcessPipeStore pipe) : Stream
    {
        public override bool CanRead => false;

        public override bool CanSeek => false;

        public override bool CanWrite => true;

        public override long Length => throw new NotSupportedException();

        public override long Position
        {
            get => throw new NotSupportedException();
            set => throw new NotSupportedException();
        }

        public override void Write(byte[] buffer, int offset, int count) => pipe.Write(buffer.AsSpan(offset, count));

        public override void Write(ReadOnlySpan<byte> buffer) => pipe.Write(buffer);

        public override ValueTask WriteAsync(ReadOnlyMemory<byte> buffer, CancellationToken cancellationToken = default)
        {
            pipe.Write(buffer.Span);
            return ValueTask.CompletedTask;
        }

        public override void Flush()
        {
        }

        public override Task FlushAsync(CancellationToken cancellationToken) => Task.CompletedTask;

        public override int Read(byte[] buffer, int offset, int count) => throw new NotSupportedException();

        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();

        public override void SetLength(long value) => throw new NotSupportedException();

        protected override void Dispose(bool disposing)
        {
            if (disposing)
            {
                pipe.CloseWriter();
            }

            base.Dispose(disposing);
        }
    }
}
