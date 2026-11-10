<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

final readonly class AddItem
{
    public function __construct(
        public string $productName,
        public int $quantity,
        public float|int $unitPrice,
        public string $currency = '',
    ) {
    }
}
