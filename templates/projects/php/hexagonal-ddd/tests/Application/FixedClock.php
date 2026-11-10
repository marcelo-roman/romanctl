<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use DateTimeImmutable;
use SampleApp\Application\Shared\Clock;

final readonly class FixedClock implements Clock
{
    public function __construct(private DateTimeImmutable $now = new DateTimeImmutable('2026-01-01T00:00:00Z'))
    {
    }

    public function now(): DateTimeImmutable
    {
        return $this->now;
    }
}
