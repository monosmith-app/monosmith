namespace MonoSmith.Plugin.Sdk.Tests.Concurrency;

public sealed class ConcurrencyTests
{
    [Theory]
    [InlineData(null, 1)]
    [InlineData(0L, 1)]
    [InlineData(1L, 1)]
    [InlineData(4L, 4)]
    [InlineData(-2L, 1)]
    [InlineData(-1L, ConcurrencyPolicy.UnlimitedInFlight)]
    public void The_host_keeps_as_many_requests_in_flight_as_the_plugin_declared(long? declared, int expected) =>
        Assert.Equal(expected, ConcurrencyPolicy.InFlightLimit(declared));

    [Theory]
    [InlineData(null, 1)]
    [InlineData(0L, 1)]
    [InlineData(3L, 3)]
    [InlineData(-5L, 1)]
    [InlineData(-1L, ConcurrencyPolicy.UnlimitedWorkers)]
    public void The_plugin_runs_as_many_workers_as_it_declared_and_unlimited_stays_bounded(long? declared, int expected) =>
        Assert.Equal(expected, ConcurrencyPolicy.WorkerLimit(declared));

    [Theory]
    [InlineData(0L, true)]
    [InlineData(1L, true)]
    [InlineData(-1L, true)]
    [InlineData(-2L, false)]
    public void Tells_a_value_the_protocol_defines_from_one_it_does_not(long declared, bool valid) =>
        Assert.Equal(valid, ConcurrencyPolicy.IsValid(declared));
}
