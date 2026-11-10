namespace SampleApp.Domain.Common;

/// <summary>Uma regra de negócio foi violada.</summary>
public sealed class DomainException(string message) : Exception(message);
