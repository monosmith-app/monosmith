using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The workspace's audit policy choices.</summary>
public static class AuditPolicies
{
    /// <summary>Whether the ecosystem's (<c>node</c>, <c>go</c>, <c>python</c>, <c>rust</c>) audit policy checks are on.</summary>
    /// <param name="audit">The workspace's audit settings, or <c>null</c>.</param>
    /// <param name="policy">The ecosystem.</param>
    /// <returns>Whether the checks run.</returns>
    public static bool IsEnabled(AuditSettings? audit, string policy) => audit?.Policies.Contains(policy) ?? false;

    /// <summary>Whether a failing check of the policy fails the audit: it does unless the workspace made the policy only warn.</summary>
    /// <param name="audit">The workspace's audit settings, or <c>null</c>.</param>
    /// <param name="policy">The ecosystem.</param>
    /// <returns>Whether a failure blocks.</returns>
    public static bool Blocks(AuditSettings? audit, string policy) => !(audit?.Warn.Contains(policy) ?? false);
}
