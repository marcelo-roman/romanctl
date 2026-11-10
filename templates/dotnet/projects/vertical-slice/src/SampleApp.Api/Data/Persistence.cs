using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Data;

internal static class Persistence
{
    public static void AddPersistence(this IHostApplicationBuilder builder)
    {
        var connectionString = builder.Configuration.GetConnectionString("Default")
            ?? throw new InvalidOperationException("Falta a connection string Default.");

        builder.Services.AddDbContext<AppDbContext>(options => options.UseNpgsql(connectionString));
    }
}
