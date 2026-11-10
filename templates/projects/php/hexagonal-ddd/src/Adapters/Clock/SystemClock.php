<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Clock;

use DateTimeImmutable;
use DateTimeZone;
use SampleApp\Application\Shared\Clock;

final class SystemClock implements Clock
{
    public function now(): DateTimeImmutable
    {
        return new DateTimeImmutable('now', new DateTimeZone('UTC'));
    }
}
