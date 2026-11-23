<?php

declare(strict_types=1);

namespace SampleApp\Tests\Domain;

use DateTimeImmutable;
use PHPUnit\Framework\TestCase;
use SampleApp\Domain\Orders\Money;
use SampleApp\Domain\Orders\Order;
use SampleApp\Domain\Orders\OrderPlaced;
use SampleApp\Domain\Orders\OrderStatus;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final class OrderTest extends TestCase
{
    private DateTimeImmutable $now;

    protected function setUp(): void
    {
        $this->now = new DateTimeImmutable('2026-01-01T12:00:00Z');
    }

    public function testCreateRequiresCustomer(): void
    {
        $this->expectException(BusinessRuleViolation::class);

        Order::create('  ', $this->now);
    }

    public function testNewOrderIsAnEmptyDraft(): void
    {
        $order = Order::create('  Ana ', $this->now);

        self::assertSame('Ana', $order->customerName());
        self::assertSame(OrderStatus::Draft, $order->status());
        self::assertSame($this->now, $order->createdAt());
        self::assertEquals(new Money(0, 'USD'), $order->total());
        self::assertSame(36, strlen($order->id()));
    }

    public function testTotalSumsItems(): void
    {
        $order = $this->draft();
        $order->addItem('Pen', 2, new Money(350, 'usd'));
        $order->addItem('Notebook', 1, new Money(1200, 'usd'));

        self::assertEquals(new Money(1900, 'USD'), $order->total());
        self::assertEquals(new Money(700, 'USD'), $order->items()[0]->subtotal());
    }

    public function testItemRequiresProduct(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('Product name is required.'));

        $this->draft()->addItem(' ', 1, new Money(100));
    }

    public function testItemRequiresPositiveQuantity(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('Quantity must be greater than zero.'));

        $this->draft()->addItem('Pen', 0, new Money(100));
    }

    public function testItemsMustShareCurrency(): void
    {
        $order = $this->draft();
        $order->addItem('Pen', 1, new Money(100, 'USD'));

        $this->expectException(BusinessRuleViolation::class);

        $order->addItem('Pen', 1, new Money(100, 'EUR'));
    }

    public function testPlaceRequiresItems(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('An order without items cannot be placed.'));

        $this->draft()->place($this->now);
    }

    public function testPlaceRaisesEventOnce(): void
    {
        $order = $this->draft();
        $order->addItem('Pen', 1, new Money(100));
        $order->place($this->now);

        self::assertSame(OrderStatus::Placed, $order->status());
        self::assertEquals([new OrderPlaced($order->id(), 100, 'USD', $this->now)], $order->pullEvents());
        self::assertSame([], $order->pullEvents());
    }

    public function testPlacedOrderCannotChange(): void
    {
        $order = $this->draft();
        $order->addItem('Pen', 1, new Money(100));
        $order->place($this->now);

        $this->assertViolation(fn () => $order->addItem('Pencil', 1, new Money(100)), 'A placed order cannot be changed.');
        $this->assertViolation(fn () => $order->place($this->now), 'The order has already been placed.');
    }

    public function testMoneyRules(): void
    {
        $this->assertViolation(fn () => new Money(-1), 'Amount cannot be negative.');
        $this->assertViolation(fn () => new Money(100, 'DOLLAR'), 'Currency must have 3 letters, like USD.');
        $this->assertViolation(fn () => (new Money(100))->add(new Money(100, 'EUR')), 'Cannot add different currencies.');

        self::assertEquals(new Money(350, 'USD'), Money::fromAmount(3.5));
        self::assertEquals(new Money(13, 'USD'), Money::fromAmount(0.125, 'usd'));
        self::assertSame(3.5, (new Money(350))->amount());
        self::assertTrue((new Money(350))->equals(Money::fromAmount(3.5, 'USD')));
    }

    public function testIdentifierParsing(): void
    {
        $id = Identifier::generate();

        self::assertSame($id, Identifier::parse(strtoupper($id)));
        $this->assertViolation(fn () => Identifier::parse('abc'), 'Invalid identifier.');
    }

    private function draft(): Order
    {
        return Order::create('Ana', $this->now);
    }

    private function assertViolation(callable $action, string $message): void
    {
        try {
            $action();
        } catch (BusinessRuleViolation $violation) {
            self::assertSame($message, $violation->getMessage());

            return;
        }
        self::fail("Expected the rule: {$message}");
    }
}
