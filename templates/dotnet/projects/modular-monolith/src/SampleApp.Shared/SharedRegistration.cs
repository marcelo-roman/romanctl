using SampleApp.Shared.Messaging;
using SampleApp.Shared.Persistence;
using SampleApp.Shared.Web;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Shared;

public static class SharedRegistration
{
    public static IHostApplicationBuilder AddShared(this IHostApplicationBuilder builder)
    {
        builder.AddModuleDatabase();
        builder.Services.AddSingleton(TimeProvider.System);
        builder.Services.AddEventBus();
        builder.Services.AddProblemDetails();
        builder.Services.AddExceptionHandler<ErrorHandling>();
        builder.Services.ConfigureHttpJsonOptions(options =>
        {
            options.SerializerOptions.RespectNullableAnnotations = true;
            options.SerializerOptions.RespectRequiredConstructorParameters = true;
        });
        return builder;
    }
}
