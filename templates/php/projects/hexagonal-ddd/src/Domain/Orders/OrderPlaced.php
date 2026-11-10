<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

use DateTimeImmutable;
use SampleApp\Domain\Shared\DomainEvent;

final readonly class OrderPlaced implements DomainEvent
{
    public function __construct(
        public string $orderId,
        public int $totalCents,
        public string $currency,
        public DateTimeImmutable $occurredAt,
    ) {
    }

    public function name(): string
    {
        return 'order.placed';
    }
}
