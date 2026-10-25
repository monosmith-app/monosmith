namespace MonoSmith.Plugin.Sdk.Tests.Cancellation;

public sealed class CancellationTests
{
    [Fact]
    public void Does_nothing_while_the_token_is_not_cancelled() =>
        CancellationToken.None.ThrowIfCancelled("never");

    [Fact]
    public void Throws_the_request_cancelled_error_with_the_message_the_answer_will_carry()
    {
        using var source = new CancellationTokenSource();
        source.Cancel();
        var error = Assert.Throws<RequestCancelledError>(() => source.Token.ThrowIfCancelled("the scan was stopped"));
        Assert.Equal("the scan was stopped", error.Message);
        Assert.Equal("the request was cancelled", Assert.Throws<RequestCancelledError>(() => source.Token.ThrowIfCancelled()).Message);
    }
}
