using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Data;

internal static class Persistence
{
    public static void AddPersistence(this IHostApplicationBuilder builder)
    {
        var connectionString = builder.Configuration.GetConnectionString("Default")
            ?? throw new InvalidOperationException("The Default connection string is missing.");

        builder.Services.AddDbContext<AppDbContext>(options => options.UseNpgsql(connectionString));
    }
}
