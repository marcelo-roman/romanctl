using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Application.{{.EntityPlural}}.Commands;

public sealed record Create{{.Entity}}(string Name) : ICommand<Guid>;

internal sealed class Create{{.Entity}}Handler(I{{.Entity}}Repository repository, IUnitOfWork unitOfWork, TimeProvider clock)
    : ICommandHandler<Create{{.Entity}}, Guid>
{
    public async Task<Guid> HandleAsync(Create{{.Entity}} command, CancellationToken cancellationToken)
    {
        var {{.EntityCamel}} = {{.Entity}}.Create(command.Name, clock.GetUtcNow());
        repository.Add({{.EntityCamel}});
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return {{.EntityCamel}}.Id.Value;
    }
}
