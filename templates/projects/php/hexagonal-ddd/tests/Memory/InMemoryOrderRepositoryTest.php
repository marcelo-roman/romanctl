<?php

declare(strict_types=1);

namespace SampleApp\Tests\Memory;

use DateTimeImmutable;
use PHPUnit\Framework\TestCase;
use SampleApp\Adapters\Memory\InMemoryOrderRepository;
use SampleApp\Adapters\Memory\MemoryStore;
use SampleApp\Domain\Orders\Money;
use SampleApp\Domain\Orders\Order;

final class InMemoryOrderRepositoryTest extends TestCase
{
    private string $directory;

    protected function setUp(): void
    {
        $this->directory = sys_get_temp_dir() . '/memory-' . bin2hex(random_bytes(4));
    }

    protected function tearDown(): void
    {
        array_map(unlink(...), glob($this->directory . '/*') ?: []);
        if (is_dir($this->directory)) {
            rmdir($this->directory);
        }
    }

    public function testStoredCopiesDoNotShareState(): void
    {
        $repository = new InMemoryOrderRepository();
        $order = Order::create('Ana', new DateTimeImmutable());
        $repository->save($order);
        $order->addItem('Caneta', 1, new Money(100));

        self::assertSame([], $repository->find($order->id())?->items());
    }

    public function testStoreKeepsOrdersBetweenInstances(): void
    {
        $order = Order::create('Ana', new DateTimeImmutable());
        $order->addItem('Caneta', 2, new Money(350));
        (new InMemoryOrderRepository(new MemoryStore($this->directory)))->save($order);

        $loaded = (new InMemoryOrderRepository(new MemoryStore($this->directory)))->find($order->id());

        self::assertNotNull($loaded);
        self::assertEquals(new Money(700), $loaded->total());
    }
}
