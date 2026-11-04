using {{.Name}}.Api;
using {{.Name}}.Application;
using {{.Name}}.Infrastructure;
{{- if .Postgres}}
using {{.Name}}.Infrastructure.Persistence.EfCore;
{{- end}}

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddApplication();
builder.Services.AddInfrastructure(builder.Configuration);
builder.Services.AddOpenApi();
builder.Services.AddProblemDetails();
builder.Services.AddExceptionHandler<ErrorHandling>();

var app = builder.Build();

app.UseExceptionHandler();
{{- if .Postgres}}

if (app.Environment.IsDevelopment())
{
    await using var scope = app.Services.CreateAsyncScope();
    await scope.ServiceProvider.GetRequiredService<AppDbContext>().Database.EnsureCreatedAsync();
}
{{- end}}

app.MapOpenApi();
app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
app.MapOrderEndpoints();

app.Run();
