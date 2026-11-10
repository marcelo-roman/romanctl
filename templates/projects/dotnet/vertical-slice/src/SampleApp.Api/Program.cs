using SampleApp.Api;
using SampleApp.Api.Data;

var builder = WebApplication.CreateBuilder(args);

builder.AddPersistence();
builder.Services.AddSingleton(TimeProvider.System);
builder.Services.AddOpenApi();
builder.Services.AddProblemDetails();
builder.Services.AddExceptionHandler<ErrorHandling>();

var app = builder.Build();

app.UseExceptionHandler();

if (app.Environment.IsDevelopment())
{
    await using var scope = app.Services.CreateAsyncScope();
    await scope.ServiceProvider.GetRequiredService<AppDbContext>().Database.EnsureCreatedAsync();
}

app.MapOpenApi();
app.MapFeatures();

app.Run();
