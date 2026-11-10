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
            throw new BusinessRuleViolation('O nome do cliente é obrigatório.');
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
            throw new BusinessRuleViolation('Um pedido já realizado não pode ser alterado.');
        }
        if ($this->items !== [] && $unitPrice->currency !== $this->currency()) {
            throw new BusinessRuleViolation('Todos os itens do pedido devem usar a mesma moeda.');
        }
        $this->items[] = OrderItem::create($productName, $quantity, $unitPrice);
    }

    public function place(DateTimeImmutable $now): void
    {
        if ($this->status !== OrderStatus::Draft) {
            throw new BusinessRuleViolation('O pedido já foi realizado.');
        }
        if ($this->items === []) {
            throw new BusinessRuleViolation('Um pedido sem itens não pode ser realizado.');
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
