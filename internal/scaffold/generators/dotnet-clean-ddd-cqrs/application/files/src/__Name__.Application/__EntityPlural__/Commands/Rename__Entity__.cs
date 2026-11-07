using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Application.{{.EntityPlural}}.Commands;

public sealed record Rename{{.Entity}}(Guid Id, string Name) : ICommand<bool>;

internal sealed class Rename{{.Entity}}Handler(I{{.Entity}}Repository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<Rename{{.Entity}}, bool>
{
    public async Task<bool> HandleAsync(Rename{{.Entity}} command, CancellationToken cancellationToken)
    {
        var {{.EntityCamel}} = await repository.GetByIdAsync(new {{.Entity}}Id(command.Id), cancellationToken)
            ?? throw new NotFoundException($"{{.Entity}} {command.Id} não encontrado.");

        {{.EntityCamel}}.Rename(command.Name);
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
