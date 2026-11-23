namespace SampleApp.Domain.Common;

/// <summary>A business rule was violated.</summary>
public sealed class DomainException(string message) : Exception(message);
