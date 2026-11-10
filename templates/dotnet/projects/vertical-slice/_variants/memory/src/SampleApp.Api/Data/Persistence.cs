using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Data;

internal static class Persistence
{
    public static void AddPersistence(this IHostApplicationBuilder builder) =>
        builder.Services.AddDbContext<AppDbContext>(options => options.UseInMemoryDatabase("sample-app"));
}
