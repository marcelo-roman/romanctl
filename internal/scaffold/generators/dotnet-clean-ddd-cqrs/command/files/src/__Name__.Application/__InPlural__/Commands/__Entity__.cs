using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.{{.InPlural}};

namespace {{.Name}}.Application.{{.InPlural}}.Commands;

public sealed record {{.Entity}}(Guid Id) : ICommand<bool>;

internal sealed class {{.Entity}}Handler(I{{.In}}Repository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<{{.Entity}}, bool>
{
    public async Task<bool> HandleAsync({{.Entity}} command, CancellationToken cancellationToken)
    {
        var {{.InCamel}} = await repository.GetByIdAsync(new {{.In}}Id(command.Id), cancellationToken)
            ?? throw new NotFoundException($"{{.In}} {command.Id} não encontrado.");

        // Chame aqui o comportamento do agregado, por exemplo: {{.InCamel}}.Archive();

        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
