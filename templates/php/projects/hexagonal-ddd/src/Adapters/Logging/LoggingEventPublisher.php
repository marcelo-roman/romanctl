<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Logging;

use Psr\Log\LoggerInterface;
use SampleApp\Application\Shared\EventPublisher;
use SampleApp\Domain\Shared\DomainEvent;

final readonly class LoggingEventPublisher implements EventPublisher
{
    public function __construct(private LoggerInterface $logger)
    {
    }

    public function publish(DomainEvent ...$events): void
    {
        foreach ($events as $event) {
            $this->logger->info('domain event name={name} event={event}', [
                'name' => $event->name(),
                'event' => json_encode($event, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
            ]);
        }
    }
}
