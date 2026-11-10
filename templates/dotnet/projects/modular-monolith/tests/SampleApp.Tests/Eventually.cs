namespace SampleApp.Tests;

public static class Eventually
{
    private static readonly TimeSpan Timeout = TimeSpan.FromSeconds(10);
    private static readonly TimeSpan Interval = TimeSpan.FromMilliseconds(50);

    public static async Task<T> MatchAsync<T>(Func<Task<T>> read, Func<T, bool> condition)
    {
        var deadline = DateTimeOffset.UtcNow + Timeout;
        var value = await read();
        while (!condition(value) && DateTimeOffset.UtcNow < deadline)
        {
            await Task.Delay(Interval);
            value = await read();
        }
        return value;
    }
}
