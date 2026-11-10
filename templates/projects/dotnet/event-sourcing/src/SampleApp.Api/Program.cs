using SampleApp.Api;
using SampleApp.Application;
using SampleApp.Infrastructure;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddApplication();
builder.Services.AddInfrastructure(builder.Configuration);
builder.Services.AddOpenApi();
builder.Services.AddProblemDetails();
builder.Services.AddExceptionHandler<ErrorHandling>();

var app = builder.Build();

app.UseExceptionHandler();

app.MapOpenApi();
app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
app.MapOrderEndpoints();
// romanctl:endpoints

app.Run();

public partial class Program;
