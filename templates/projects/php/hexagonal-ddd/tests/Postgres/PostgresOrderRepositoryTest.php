<?php

declare(strict_types=1);

namespace SampleApp\Tests\Postgres;

use DateTimeImmutable;
use DateTimeZone;
use SampleApp\Adapters\Postgres\PostgresOrderRepository;
use SampleApp\Domain\Orders\Money;
use SampleApp\Domain\Orders\Order;
use SampleApp\Domain\Orders\OrderStatus;
use SampleApp\Domain\Shared\Identifier;

final class PostgresOrderRepositoryTest extends DatabaseTestCase
{
    public function testOrderRoundTrip(): void
    {
        $repository = new PostgresOrderRepository($this->database);
        $order = Order::create('Ana', new DateTimeImmutable('now', new DateTimeZone('UTC')));
        $order->addItem('Caneta', 2, new Money(350));
        $order->addItem('Caderno', 1, new Money(1200));
        $repository->save($order);

        $loaded = $repository->find($order->id());

        self::assertNotNull($loaded);
        self::assertSame('Ana', $loaded->customerName());
        self::assertEquals($order->createdAt(), $loaded->createdAt());
        self::assertEquals(new Money(1900), $loaded->total());
        self::assertSame(['Caneta', 'Caderno'], array_map(static fn ($item) => $item->productName, $loaded->items()));
        self::assertNull($repository->find(Identifier::generate()));
    }

    public function testSavingAgainUpdatesStatusAndItems(): void
    {
        $repository = new PostgresOrderRepository($this->database);
        $order = Order::create('Bia', new DateTimeImmutable('now', new DateTimeZone('UTC')));
        $repository->save($order);
        $order->addItem('Lápis', 3, new Money(100));
        $order->place(new DateTimeImmutable());
        $repository->save($order);

        $loaded = $repository->find($order->id());

        self::assertNotNull($loaded);
        self::assertSame(OrderStatus::Placed, $loaded->status());
        self::assertEquals(new Money(300), $loaded->total());
        self::assertContains($order->id(), array_map(static fn (Order $recent) => $recent->id(), $repository->listRecent(200)));
    }

    public function testSchemaCreationIsIdempotent(): void
    {
        $this->database->connection();

        self::assertSame(1, (int) $this->newDatabase()->connection()->fetchOne('SELECT 1'));
    }
}
