using System.Security.Cryptography;
using System.Text;

namespace Monosmith.Launcher.SignatureCheck;

/// <summary>Checks a downloaded archive against smith's signed checksums.txt (minisign).</summary>
internal static class VerifyChecksumsAlgorithm
{
    /// <summary>
    /// The public key smith's releases are signed with (minisign, key ID 6B2CE84DD9F901). Its private half is the
    /// repository secret SMITH_SIGNING_KEY (ADR 0101).
    /// </summary>
    internal const string ReleasePublicKey = "RWQB+dlN6CxrAMNDX47n/BQK3mOTgrJ4yVIasbrbP6xRRZTQeB5yncav";

    private const string Trusted = "trusted comment: ";

    /// <summary>
    /// Checks a minisign signature (the plain Ed25519 kind smith release writes) of a message: the key ID, the signature
    /// of the message, and the signature of the trusted comment.
    /// </summary>
    /// <param name="message">The signed file's bytes (checksums.txt).</param>
    /// <param name="signatureText">The .minisig file's text.</param>
    /// <param name="publicKeyText">The public key in base64, as minisign prints it.</param>
    /// <returns>True only when every check passes.</returns>
    internal static bool VerifyMinisign(byte[] message, string signatureText, string publicKeyText)
    {
        var lines = signatureText.Split('\n').Select(l => l.TrimEnd('\r')).Where(l => l.Length > 0).ToList();
        var signatureLine = lines.Find(l => !l.StartsWith("untrusted comment:", StringComparison.Ordinal) && !l.StartsWith("trusted comment:", StringComparison.Ordinal));
        var trusted = lines.Find(l => l.StartsWith(Trusted, StringComparison.Ordinal));
        if (signatureLine is null || trusted is null || lines[^1] == signatureLine)
        {
            return false;
        }
        byte[] sig, pub, global;
        try
        {
            sig = Convert.FromBase64String(signatureLine);
            pub = Convert.FromBase64String(publicKeyText.Trim());
            global = Convert.FromBase64String(lines[^1]);
        }
        catch (FormatException)
        {
            return false;
        }
        // Ed + key ID (8) + signature (64); Ed + key ID (8) + key (32).
        if (sig.Length != 74 || pub.Length != 42 || sig[0] != 'E' || sig[1] != 'd' || !sig.AsSpan(2, 8).SequenceEqual(pub.AsSpan(2, 8)))
        {
            return false;
        }
        var key = pub[10..];
        var messageSignature = sig[10..];
        if (!Ed25519Algorithm.Verify(key, message, messageSignature))
        {
            return false;
        }
        byte[] comment = [.. messageSignature, .. Encoding.UTF8.GetBytes(trusted[Trusted.Length..])];
        return Ed25519Algorithm.Verify(key, comment, global);
    }

    /// <summary>The SHA-256 a checksums.txt lists for a file.</summary>
    /// <param name="checksums">The text of checksums.txt (<c>&lt;sha256&gt;  &lt;name&gt;</c> lines).</param>
    /// <param name="name">The file name.</param>
    /// <returns>The hex digest, or null when the file is not listed.</returns>
    internal static string? ListedDigest(string checksums, string name)
    {
        foreach (var line in checksums.Split('\n'))
        {
            var parts = line.Trim().Split((char[]?)null, 2, StringSplitOptions.RemoveEmptyEntries);
            if (parts.Length == 2 && parts[1].TrimStart('*') == name)
            {
                return parts[0].ToLowerInvariant();
            }
        }
        return null;
    }

    /// <summary>Checks a downloaded archive against a signed checksums.txt.</summary>
    /// <param name="archive">The archive's bytes.</param>
    /// <param name="name">The archive's file name.</param>
    /// <param name="checksums">checksums.txt.</param>
    /// <param name="signature">checksums.txt.minisig.</param>
    /// <param name="publicKey">The public key; default the release key.</param>
    /// <returns>Why the archive is not trusted, or null when it is.</returns>
    internal static string? CheckArchive(byte[] archive, string name, byte[] checksums, string signature, string publicKey = ReleasePublicKey)
    {
        if (!VerifyMinisign(checksums, signature, publicKey))
        {
            return "the signature of checksums.txt does not match smith's release key";
        }
        var listed = ListedDigest(Encoding.UTF8.GetString(checksums), name);
        if (listed is null)
        {
            return $"checksums.txt does not list {name}";
        }
        var actual = Convert.ToHexString(SHA256.HashData(archive)).ToLowerInvariant();
        return actual == listed ? null : $"{name} does not match its checksum (got {actual}, listed {listed})";
    }
}
