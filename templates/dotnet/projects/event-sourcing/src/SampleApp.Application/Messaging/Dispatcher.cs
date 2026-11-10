using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Application.Messaging;

internal sealed class Dispatcher(IServiceProvider services) : IDispatcher
{
    public Task<TResult> SendAsync<TResult>(ICommand<TResult> command, CancellationToken cancellationToken = default) =>
        Invoke<TResult>(typeof(ICommandHandler<,>), command, cancellationToken);

    public Task<TResult> QueryAsync<TResult>(IQuery<TResult> query, CancellationToken cancellationToken = default) =>
        Invoke<TResult>(typeof(IQueryHandler<,>), query, cancellationToken);

    private Task<TResult> Invoke<TResult>(Type openHandlerType, object message, CancellationToken cancellationToken)
    {
        var handlerType = openHandlerType.MakeGenericType(message.GetType(), typeof(TResult));
        var handler = services.GetRequiredService(handlerType);
        return (Task<TResult>)handlerType.GetMethod("HandleAsync")!.Invoke(handler, [message, cancellationToken])!;
    }
}
