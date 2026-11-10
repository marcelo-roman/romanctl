<?php

declare(strict_types=1);

namespace SampleApp\Application\Shared;

use DateTimeImmutable;

interface Clock
{
    public function now(): DateTimeImmutable;
}
