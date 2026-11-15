using System.Numerics;
using System.Security.Cryptography;

namespace Monosmith.Launcher.SignatureCheck;

/// <summary>Ed25519 signature checking (RFC 8032, section 5.1.7), without a dependency.</summary>
/// <remarks>
/// .NET has no Ed25519, and the tool installs with nothing else. It only checks public signatures, so it holds no
/// secret that its timing could reveal.
/// </remarks>
internal static class Ed25519Algorithm
{
    private static readonly BigInteger P = BigInteger.Pow(2, 255) - 19;
    private static readonly BigInteger L = BigInteger.Pow(2, 252) + BigInteger.Parse("27742317777372353535851937790883648493", System.Globalization.CultureInfo.InvariantCulture);
    private static readonly BigInteger D = Mod(-121665 * Inverse(121666));
    private static readonly BigInteger SqrtM1 = BigInteger.ModPow(2, (P - 1) / 4, P);
    private static readonly Point? Base = Decode(Convert.FromHexString("5866666666666666666666666666666666666666666666666666666666666666"));

    private readonly record struct Point(BigInteger X, BigInteger Y, BigInteger Z, BigInteger T);

    /// <summary>Checks an Ed25519 signature.</summary>
    /// <param name="publicKey">The 32 key bytes.</param>
    /// <param name="message">The signed bytes.</param>
    /// <param name="signature">The 64 signature bytes.</param>
    /// <returns>True only when the signature is the key's signature of the message.</returns>
    internal static bool Verify(byte[] publicKey, byte[] message, byte[] signature)
    {
        if (signature.Length != 64 || Base is not { } basePoint)
        {
            return false;
        }
        var a = Decode(publicKey);
        var r = Decode(signature[..32]);
        var s = new BigInteger(signature.AsSpan(32), isUnsigned: true);
        if (a is not { } key || r is not { } rPoint || s >= L)
        {
            return false;
        }
        var h = Mod(new BigInteger(SHA512.HashData([.. signature[..32], .. publicKey, .. message]), isUnsigned: true), L);
        return Same(Multiply(s, basePoint), Add(rPoint, Multiply(h, key)));
    }

    private static BigInteger Mod(BigInteger x) => Mod(x, P);

    private static BigInteger Mod(BigInteger x, BigInteger m) => ((x % m) + m) % m;

    private static BigInteger Inverse(BigInteger x) => BigInteger.ModPow(Mod(x), P - 2, P);

    private static Point Add(Point a, Point b)
    {
        var k = Mod((a.Y - a.X) * (b.Y - b.X));
        var m = Mod((a.Y + a.X) * (b.Y + b.X));
        var c = Mod(a.T * 2 * D * b.T);
        var d = Mod(a.Z * 2 * b.Z);
        BigInteger e = m - k, f = d - c, g = d + c, h = m + k;
        return new Point(Mod(e * f), Mod(g * h), Mod(f * g), Mod(e * h));
    }

    private static Point Multiply(BigInteger s, Point point)
    {
        var result = new Point(0, 1, 1, 0);
        while (s > 0)
        {
            if (!s.IsEven)
            {
                result = Add(result, point);
            }
            point = Add(point, point);
            s >>= 1;
        }
        return result;
    }

    private static bool Same(Point a, Point b) => Mod((a.X * b.Z) - (b.X * a.Z)) == 0 && Mod((a.Y * b.Z) - (b.Y * a.Z)) == 0;

    private static Point? Decode(byte[] data)
    {
        if (data.Length != 32)
        {
            return null;
        }
        var y = new BigInteger(data, isUnsigned: true);
        var sign = (int)(y >> 255);
        y &= (BigInteger.One << 255) - 1;
        if (y >= P)
        {
            return null;
        }
        var x2 = Mod((y * y - 1) * Inverse((D * y * y) + 1));
        if (x2.IsZero)
        {
            return sign == 1 ? null : new Point(0, y, 1, 0);
        }
        var x = BigInteger.ModPow(x2, (P + 3) / 8, P);
        if (!Mod((x * x) - x2).IsZero)
        {
            x = Mod(x * SqrtM1);
        }
        if (!Mod((x * x) - x2).IsZero)
        {
            return null;
        }
        if ((x.IsEven ? 0 : 1) != sign)
        {
            x = P - x;
        }
        return new Point(x, y, 1, Mod(x * y));
    }
}
