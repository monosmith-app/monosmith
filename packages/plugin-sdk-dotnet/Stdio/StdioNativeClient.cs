using System.ComponentModel;
using System.Runtime.InteropServices;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// The operating system calls that take the standard handles: kernel32 on Windows, libc elsewhere. The .NET
/// runtime has no way to duplicate or replace the process's standard handles.
/// </summary>
internal static class StdioNativeClient
{
    public const int StdInput = -10;
    public const int StdOutput = -11;
    public const int StdError = -12;

    private const uint DuplicateSameAccess = 2;
    private const uint GenericRead = 0x80000000;
    private const uint FileShareReadWrite = 3;
    private const uint OpenExisting = 3;
    private const int FSetFd = 2;
    private const int FdCloexec = 1;

    /// <summary>Whether <c>fcntl</c> can be called as declared here: not on Apple ARM64, which passes variadic arguments differently.</summary>
    public static bool CanSetCloseOnExec =>
        !(OperatingSystem.IsMacOS() && RuntimeInformation.ProcessArchitecture == Architecture.Arm64);

    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern nint GetStdHandle(int standardHandle);

    [DllImport("kernel32.dll", SetLastError = true)]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool SetStdHandle(int standardHandle, nint handle);

    /// <summary>A copy of a standard handle that a child process does not inherit.</summary>
    public static nint Duplicate(nint handle)
    {
        var self = GetCurrentProcess();
        if (!DuplicateHandle(self, handle, self, out var copy, 0, false, DuplicateSameAccess))
        {
            throw new IOException("duplicating a standard handle: " + new Win32Exception(Marshal.GetLastWin32Error()).Message);
        }

        return copy;
    }

    /// <summary>The null device, as a handle a child process inherits.</summary>
    public static nint OpenNullDeviceWindows()
    {
        var attributes = new SecurityAttributes { Length = Marshal.SizeOf<SecurityAttributes>(), InheritHandle = 1 };
        var handle = CreateFileW("NUL", GenericRead, FileShareReadWrite, ref attributes, OpenExisting, 0, 0);
        if (handle == -1)
        {
            throw new IOException("opening the null device: " + new Win32Exception(Marshal.GetLastWin32Error()).Message);
        }

        return handle;
    }

    public static void ThrowIfFailed(bool succeeded, string what)
    {
        if (!succeeded)
        {
            throw new IOException($"{what}: {new Win32Exception(Marshal.GetLastWin32Error()).Message}");
        }
    }

    public static int Dup(int descriptor) => Checked(dup(descriptor), "duplicating a standard stream");

    public static void Dup2(int source, int target) => Checked(dup2(source, target), "redirecting a standard stream");

    public static void Close(int descriptor) => Checked(close(descriptor), "closing a descriptor");

    public static int OpenNullDeviceUnix() => Checked(open("/dev/null", 0), "opening the null device");

    public static void SetCloseOnExec(int descriptor) => Checked(fcntl(descriptor, FSetFd, FdCloexec), "marking a descriptor close-on-exec");

    private static int Checked(int result, string what)
    {
        if (result < 0)
        {
            throw new IOException($"{what}: {new Win32Exception(Marshal.GetLastWin32Error()).Message}");
        }

        return result;
    }

    [DllImport("kernel32.dll")]
    private static extern nint GetCurrentProcess();

    [DllImport("kernel32.dll", SetLastError = true)]
    [return: MarshalAs(UnmanagedType.Bool)]
    private static extern bool DuplicateHandle(nint sourceProcess, nint source, nint targetProcess, out nint target, uint access, [MarshalAs(UnmanagedType.Bool)] bool inherit, uint options);

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern nint CreateFileW(string name, uint access, uint share, ref SecurityAttributes attributes, uint creation, uint flags, nint template);

    [DllImport("libc", SetLastError = true)]
    private static extern int dup(int descriptor);

    [DllImport("libc", SetLastError = true)]
    private static extern int dup2(int source, int target);

    [DllImport("libc", SetLastError = true)]
    private static extern int close(int descriptor);

    [DllImport("libc", SetLastError = true)]
    private static extern int open([MarshalAs(UnmanagedType.LPStr)] string path, int flags);

    [DllImport("libc", SetLastError = true)]
    private static extern int fcntl(int descriptor, int command, int argument);

    [StructLayout(LayoutKind.Sequential)]
    private struct SecurityAttributes
    {
        public int Length;
        public nint SecurityDescriptor;
        public int InheritHandle;
    }
}
