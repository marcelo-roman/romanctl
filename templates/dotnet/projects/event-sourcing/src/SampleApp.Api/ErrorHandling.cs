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
            BadHttpRequestException => (StatusCodes.Status400BadRequest, "Invalid request"),
            NotFoundException => (StatusCodes.Status404NotFound, "Not found"),
            ConcurrencyException => (StatusCodes.Status409Conflict, "Version conflict"),
            DomainException => (StatusCodes.Status422UnprocessableEntity, "Business rule violated"),
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
