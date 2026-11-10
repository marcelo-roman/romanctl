<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

use Closure;
use SampleApp\Application\Shared\Clock;
use SampleApp\Application\Shared\EventPublisher;
use SampleApp\Application\Shared\ListLimit;
use SampleApp\Domain\Orders\Money;
use SampleApp\Domain\Orders\Order;
use SampleApp\Domain\Shared\Identifier;

final readonly class OrderService
{
    public function __construct(
        private OrderRepository $repository,
        private EventPublisher $events,
        private Clock $clock,
    ) {
    }

    public function create(string $customerName): string
    {
        $order = Order::create($customerName, $this->clock->now());
        $this->repository->save($order);

        return $order->id();
    }

    public function addItem(string $orderId, AddItem $command): void
    {
        $price = Money::fromAmount($command->unitPrice, $command->currency);
        $this->change(
            $orderId,
            static fn (Order $order) => $order->addItem($command->productName, $command->quantity, $price),
        );
    }

    public function place(string $orderId): void
    {
        $this->change($orderId, fn (Order $order) => $order->place($this->clock->now()));
    }

    public function get(string $orderId): OrderView
    {
        return OrderView::of($this->load($orderId));
    }

    public function listRecent(int $limit = ListLimit::DEFAULT): array
    {
        return array_map(OrderView::of(...), $this->repository->listRecent(ListLimit::bound($limit)));
    }

    private function load(string $orderId): Order
    {
        $order = $this->repository->find(Identifier::parse($orderId));
        if ($order === null) {
            throw new OrderNotFound();
        }

        return $order;
    }

    private function change(string $orderId, Closure $apply): void
    {
        $order = $this->load($orderId);
        $apply($order);
        $this->repository->save($order);
        $this->events->publish(...$order->pullEvents());
    }
}
