using Microsoft.Win32.SafeHandles;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Taking the process's standard input and output for the protocol and nothing else.</summary>
public static class ClaimStdioUseCase
{
    /// <summary>
    /// Takes standard input and output for the protocol, and leaves nothing else a way to them.
    /// </summary>
    /// <remarks>
    /// <para>
    /// A <c>Console.WriteLine</c>, a library's write to standard output or a child process that inherits the
    /// plugin's standard output would corrupt the stream the host reads. So the protocol gets private copies of the
    /// two handles, and from here on:
    /// </para>
    /// <list type="bullet">
    /// <item><description>standard output, and <see cref="Console.Out"/>, are standard error: whatever else writes to standard output reaches the host's log, not the stream;</description></item>
    /// <item><description>standard input is the null device: a child process that reads standard input does not take the host's requests.</description></item>
    /// </list>
    /// <para>The returned streams are the only way to the real ones. Call it once, at the start of the process.</para>
    /// </remarks>
    /// <returns>The streams of the protocol.</returns>
    public static ProtocolStdio Claim()
    {
        Console.Out.Flush();
        Console.Error.Flush();
        var (input, output) = OperatingSystem.IsWindows() ? ClaimWindows() : ClaimUnix();
        Console.SetOut(Console.Error);
        Console.SetIn(TextReader.Null);
#pragma warning disable CA2000 // the streams own the handles
        return new ProtocolStdio(
            new FileStream(new SafeFileHandle(input, ownsHandle: true), FileAccess.Read, bufferSize: 1, isAsync: false),
            new FileStream(new SafeFileHandle(output, ownsHandle: true), FileAccess.Write, bufferSize: 1, isAsync: false));
#pragma warning restore CA2000
    }

    private static (nint Input, nint Output) ClaimWindows()
    {
        var stdin = StdioNativeClient.GetStdHandle(StdioNativeClient.StdInput);
        var stdout = StdioNativeClient.GetStdHandle(StdioNativeClient.StdOutput);
        var stderr = StdioNativeClient.GetStdHandle(StdioNativeClient.StdError);
        var input = StdioNativeClient.Duplicate(stdin);
        var output = StdioNativeClient.Duplicate(stdout);
        StdioNativeClient.ThrowIfFailed(StdioNativeClient.SetStdHandle(StdioNativeClient.StdOutput, stderr), "redirecting standard output");
        var none = StdioNativeClient.OpenNullDeviceWindows();
        StdioNativeClient.ThrowIfFailed(StdioNativeClient.SetStdHandle(StdioNativeClient.StdInput, none), "redirecting standard input");
        return (input, output);
    }

    private static (nint Input, nint Output) ClaimUnix()
    {
        var input = StdioNativeClient.Dup(0);
        var output = StdioNativeClient.Dup(1);

        // The copies must not leak into a child process the plugin starts, or the host would not see the end of the
        // stream while that child lives. fcntl is variadic, which the Apple ARM64 calling convention passes on the
        // stack, where this declaration does not put it: there the copies are left inheritable.
        if (StdioNativeClient.CanSetCloseOnExec)
        {
            StdioNativeClient.SetCloseOnExec(input);
            StdioNativeClient.SetCloseOnExec(output);
        }

        StdioNativeClient.Dup2(2, 1);
        var none = StdioNativeClient.OpenNullDeviceUnix();
        StdioNativeClient.Dup2(none, 0);
        StdioNativeClient.Close(none);
        return (input, output);
    }
}
