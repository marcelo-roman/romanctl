using SampleApp.Shared.Application;
using SampleApp.Shared.Domain;
using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Shared.Web;

internal sealed class ErrorHandling(IProblemDetailsService problemDetails) : IExceptionHandler
{
    public async ValueTask<bool> TryHandleAsync(HttpContext context, Exception exception, CancellationToken cancellationToken)
    {
        var (status, title) = exception switch
        {
            NotFoundException => (StatusCodes.Status404NotFound, "Not found"),
            DomainException => (StatusCodes.Status422UnprocessableEntity, "Business rule violated"),
            BadHttpRequestException badRequest => (badRequest.StatusCode, "Invalid request"),
            _ => (0, ""),
        };
        if (status == 0)
            return false;

        context.Response.StatusCode = status;
        return await problemDetails.TryWriteAsync(new ProblemDetailsContext
        {
            HttpContext = context,
            Exception = exception,
            ProblemDetails = new ProblemDetails { Status = status, Title = title, Detail = exception.Message },
        });
    }
}
