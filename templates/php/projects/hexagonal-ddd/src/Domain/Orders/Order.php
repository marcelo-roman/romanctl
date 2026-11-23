<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

use DateTimeImmutable;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final class Order
{
    private array $items;
    private array $events = [];

    private function __construct(
        private readonly string $id,
        private readonly string $customerName,
        private OrderStatus $status,
        private readonly DateTimeImmutable $createdAt,
        OrderItem ...$items,
    ) {
        $this->items = $items;
    }

    public static function create(string $customerName, DateTimeImmutable $now): self
    {
        if (trim($customerName) === '') {
            throw new BusinessRuleViolation('Customer name is required.');
        }

        return new self(Identifier::generate(), trim($customerName), OrderStatus::Draft, $now);
    }

    public static function restore(
        string $id,
        string $customerName,
        OrderStatus $status,
        DateTimeImmutable $createdAt,
        OrderItem ...$items,
    ): self {
        return new self($id, $customerName, $status, $createdAt, ...$items);
    }

    public function id(): string
    {
        return $this->id;
    }

    public function customerName(): string
    {
        return $this->customerName;
    }

    public function status(): OrderStatus
    {
        return $this->status;
    }

    public function createdAt(): DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function items(): array
    {
        return $this->items;
    }

    public function currency(): string
    {
        if ($this->items === []) {
            return Money::DEFAULT_CURRENCY;
        }

        return $this->items[0]->unitPrice->currency;
    }

    public function total(): Money
    {
        $total = Money::zero($this->currency());
        foreach ($this->items as $item) {
            $total = $total->add($item->subtotal());
        }

        return $total;
    }

    public function addItem(string $productName, int $quantity, Money $unitPrice): void
    {
        if ($this->status !== OrderStatus::Draft) {
            throw new BusinessRuleViolation('A placed order cannot be changed.');
        }
        if ($this->items !== [] && $unitPrice->currency !== $this->currency()) {
            throw new BusinessRuleViolation('All order items must use the same currency.');
        }
        $this->items[] = OrderItem::create($productName, $quantity, $unitPrice);
    }

    public function place(DateTimeImmutable $now): void
    {
        if ($this->status !== OrderStatus::Draft) {
            throw new BusinessRuleViolation('The order has already been placed.');
        }
        if ($this->items === []) {
            throw new BusinessRuleViolation('An order without items cannot be placed.');
        }
        $this->status = OrderStatus::Placed;
        $total = $this->total();
        $this->events[] = new OrderPlaced($this->id, $total->cents, $total->currency, $now);
    }

    public function pullEvents(): array
    {
        $events = $this->events;
        $this->events = [];

        return $events;
    }
}
