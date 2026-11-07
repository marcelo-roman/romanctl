using {{.Name}}.Domain.Common;
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Tests;

public class {{.Entity}}Tests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Criar_levanta_evento()
    {
        var {{.EntityCamel}} = {{.Entity}}.Create(" Exemplo ", Now);

        Assert.Equal("Exemplo", {{.EntityCamel}}.Name);
        Assert.IsType<{{.Entity}}Created>(Assert.Single({{.EntityCamel}}.DomainEvents));
    }

    [Fact]
    public void Rejeita_nome_vazio() =>
        Assert.Throws<DomainException>(() => {{.Entity}}.Create(" ", Now));
}
