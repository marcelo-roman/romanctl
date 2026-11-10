<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

use SampleApp\Domain\Orders\OrderItem;

final readonly class OrderItemView
{
    public function __construct(
        public string $productName,
        public int $quantity,
        public float $unitPrice,
        public float $subtotal,
    ) {
    }

    public static function of(OrderItem $item): self
    {
        return new self($item->productName, $item->quantity, $item->unitPrice->amount(), $item->subtotal()->amount());
    }
}
