using SampleApp.Application.Common;
using SampleApp.Domain.Common;
using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Api;

internal sealed class ErrorHandling : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(HttpContext context, Exception exception, CancellationToken cancellationToken)
    {
        var (status, title) = exception switch
        {
            BadHttpRequestException => (StatusCodes.Status400BadRequest, "Requisição inválida"),
            NotFoundException => (StatusCodes.Status404NotFound, "Não encontrado"),
            ConcurrencyException => (StatusCodes.Status409Conflict, "Conflito de versão"),
            DomainException => (StatusCodes.Status422UnprocessableEntity, "Regra de negócio violada"),
            _ => (0, ""),
        };
        if (status == 0)
            return false;

        context.Response.StatusCode = status;
        await context.Response.WriteAsJsonAsync(
            new ProblemDetails { Status = status, Title = title, Detail = exception.Message },
            options: null,
            contentType: "application/problem+json",
            cancellationToken);
        return true;
    }
}
