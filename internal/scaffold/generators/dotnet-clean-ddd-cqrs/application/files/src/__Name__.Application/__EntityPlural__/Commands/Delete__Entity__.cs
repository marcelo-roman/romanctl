using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Application.{{.EntityPlural}}.Commands;

public sealed record Delete{{.Entity}}(Guid Id) : ICommand<bool>;

internal sealed class Delete{{.Entity}}Handler(I{{.Entity}}Repository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<Delete{{.Entity}}, bool>
{
    public async Task<bool> HandleAsync(Delete{{.Entity}} command, CancellationToken cancellationToken)
    {
        var {{.EntityCamel}} = await repository.GetByIdAsync(new {{.Entity}}Id(command.Id), cancellationToken)
            ?? throw new NotFoundException($"{{.Entity}} {command.Id} não encontrado.");

        repository.Remove({{.EntityCamel}});
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
