namespace SampleApp.Tests.Postgres;

public sealed class PostgresFactAttribute : FactAttribute
{
    public const string Variable = "ConnectionStrings__Default";

    public PostgresFactAttribute()
    {
        if (string.IsNullOrWhiteSpace(Environment.GetEnvironmentVariable(Variable)))
            Skip = $"Defina {Variable} para rodar contra o PostgreSQL.";
    }
}
