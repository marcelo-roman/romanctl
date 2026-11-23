using SampleApp.Application.Common;
using SampleApp.Domain.Common;
using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Api;

/// <summary>Maps expected errors to HTTP responses. Everything else stays a 500 error.</summary>
internal sealed class ErrorHandling : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(HttpContext context, Exception exception, CancellationToken cancellationToken)
    {
        var (status, title) = exception switch
        {
            NotFoundException => (StatusCodes.Status404NotFound, "Not found"),
            DomainException => (StatusCodes.Status422UnprocessableEntity, "Business rule violated"),
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
