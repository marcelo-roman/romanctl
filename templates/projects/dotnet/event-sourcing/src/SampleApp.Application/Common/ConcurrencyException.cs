namespace SampleApp.Application.Common;

public sealed class ConcurrencyException(Guid streamId, Exception? inner = null)
    : Exception($"O registro {streamId} foi alterado por outra operação. Carregue de novo e repita.", inner);
