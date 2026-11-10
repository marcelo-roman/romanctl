<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

use DateTimeImmutable;
use SampleApp\Domain\Orders\Order;

final readonly class OrderView
{
    public array $items;

    public function __construct(
        public string $id,
        public string $customerName,
        public string $status,
        public DateTimeImmutable $createdAt,
        public float $total,
        public string $currency,
        OrderItemView ...$items,
    ) {
        $this->items = $items;
    }

    public static function of(Order $order): self
    {
        $total = $order->total();

        return new self(
            $order->id(),
            $order->customerName(),
            $order->status()->value,
            $order->createdAt(),
            $total->amount(),
            $total->currency,
            ...array_map(OrderItemView::of(...), $order->items()),
        );
    }
}
