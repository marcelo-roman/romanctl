<?php

declare(strict_types=1);

namespace SampleApp\Domain\Shared;

interface DomainEvent
{
    public function name(): string;
}
