namespace SampleApp.Shared.Application;

public sealed class NotFoundException(string message) : Exception(message);
