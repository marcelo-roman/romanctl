<?php

declare(strict_types=1);

namespace SampleApp\Application\Shared;

use SampleApp\Domain\Shared\DomainEvent;

interface EventPublisher
{
    public function publish(DomainEvent ...$events): void;
}
