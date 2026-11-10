<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use SampleApp\Application\Orders\OrderRepository;
use SampleApp\Domain\Orders\Order;

final class FakeOrderRepository implements OrderRepository
{
    public array $saved = [];
    public array $limits = [];

    public function save(Order $order): void
    {
        $this->saved[$order->id()] = $order;
    }

    public function find(string $id): ?Order
    {
        return $this->saved[$id] ?? null;
    }

    public function listRecent(int $limit): array
    {
        $this->limits[] = $limit;

        return array_slice(array_values($this->saved), 0, $limit);
    }
}
