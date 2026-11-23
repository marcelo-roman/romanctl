namespace SampleApp.Application.Common;

public sealed class ConcurrencyException(Guid streamId, Exception? inner = null)
    : Exception($"Record {streamId} was changed by another operation. Load it again and retry.", inner);
