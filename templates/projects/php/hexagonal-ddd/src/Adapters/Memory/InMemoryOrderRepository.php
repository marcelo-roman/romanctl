<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Memory;

use SampleApp\Application\Orders\OrderRepository;
use SampleApp\Domain\Orders\Order;

final class InMemoryOrderRepository implements OrderRepository
{
    private const string COLLECTION = 'orders';

    private array $orders;

    public function __construct(private readonly MemoryStore $store = new MemoryStore())
    {
        $this->orders = $store->load(self::COLLECTION);
    }

    public function save(Order $order): void
    {
        $this->orders[$order->id()] = self::copy($order);
        $this->store->save(self::COLLECTION, $this->orders);
    }

    public function find(string $id): ?Order
    {
        if (!isset($this->orders[$id])) {
            return null;
        }

        return self::copy($this->orders[$id]);
    }

    public function listRecent(int $limit): array
    {
        $orders = array_reverse(array_values($this->orders));
        usort($orders, static fn (Order $a, Order $b) => $b->createdAt() <=> $a->createdAt());

        return array_map(self::copy(...), array_slice($orders, 0, $limit));
    }

    private static function copy(Order $order): Order
    {
        return Order::restore($order->id(), $order->customerName(), $order->status(), $order->createdAt(), ...$order->items());
    }
}
