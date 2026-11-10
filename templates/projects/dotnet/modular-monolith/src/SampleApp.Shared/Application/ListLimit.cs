namespace SampleApp.Shared.Application;

public static class ListLimit
{
    public const int Default = 50;
    public const int Max = 200;

    public static int Clamp(int? limit) => Math.Clamp(limit ?? Default, 1, Max);
}
