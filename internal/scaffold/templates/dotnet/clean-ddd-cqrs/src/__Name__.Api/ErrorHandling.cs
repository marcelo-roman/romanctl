using {{.Name}}.Application.Common;
using {{.Name}}.Domain.Common;
using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace {{.Name}}.Api;

/// <summary>Traduz erros esperados em respostas HTTP. Os demais continuam sendo erro 500.</summary>
internal sealed class ErrorHandling : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(HttpContext context, Exception exception, CancellationToken cancellationToken)
    {
        var (status, title) = exception switch
        {
            NotFoundException => (StatusCodes.Status404NotFound, "Não encontrado"),
            DomainException => (StatusCodes.Status422UnprocessableEntity, "Regra de negócio violada"),
            _ => (0, ""),
        };
        if (status == 0)
            return false;

        context.Response.StatusCode = status;
        await context.Response.WriteAsJsonAsync(
            new ProblemDetails { Status = status, Title = title, Detail = exception.Message },
            cancellationToken);
        return true;
    }
}
