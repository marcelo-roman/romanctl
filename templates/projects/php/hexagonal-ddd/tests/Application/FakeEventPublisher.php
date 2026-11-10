<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use SampleApp\Application\Shared\EventPublisher;
use SampleApp\Domain\Shared\DomainEvent;

final class FakeEventPublisher implements EventPublisher
{
    public array $published = [];

    public function publish(DomainEvent ...$events): void
    {
        array_push($this->published, ...$events);
    }
}
