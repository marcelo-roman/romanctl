using SampleApp.Api;
using SampleApp.Shared;
using SampleApp.Shared.Persistence;

var builder = WebApplication.CreateBuilder(args);

builder.AddShared();
builder.AddModules();
builder.Services.AddOpenApi();

var app = builder.Build();

app.UseExceptionHandler();

if (app.Environment.IsDevelopment())
    await app.Services.InitializeModuleDatabasesAsync();

app.MapOpenApi();
app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
app.MapModules();

app.Run();
