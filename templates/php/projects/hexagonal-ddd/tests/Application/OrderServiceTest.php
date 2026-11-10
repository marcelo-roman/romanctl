<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use PHPUnit\Framework\TestCase;
use SampleApp\Application\Orders\AddItem;
use SampleApp\Application\Orders\OrderNotFound;
use SampleApp\Application\Orders\OrderService;
use SampleApp\Application\Shared\ListLimit;
use SampleApp\Domain\Orders\OrderPlaced;
use SampleApp\Domain\Orders\OrderStatus;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final class OrderServiceTest extends TestCase
{
    private FakeOrderRepository $repository;
    private FakeEventPublisher $events;
    private FixedClock $clock;
    private OrderService $service;

    protected function setUp(): void
    {
        $this->repository = new FakeOrderRepository();
        $this->events = new FakeEventPublisher();
        $this->clock = new FixedClock();
        $this->service = new OrderService($this->repository, $this->events, $this->clock);
    }

    public function testPlacingPublishesEvent(): void
    {
        $id = $this->service->create('Ana');
        $this->service->addItem($id, new AddItem('Caneta', 3, 2.5));
        $this->service->place($id);

        $view = $this->service->get($id);

        self::assertSame(['Placed', 7.5, 'BRL'], [$view->status, $view->total, $view->currency]);
        self::assertEquals([new OrderPlaced($id, 750, 'BRL', $this->clock->now())], $this->events->published);
    }

    public function testExplicitCurrencyIsKept(): void
    {
        $id = $this->service->create('Ana');
        $this->service->addItem($id, new AddItem('Pen', 1, 1, 'usd'));

        self::assertSame('USD', $this->service->get($id)->currency);
    }

    public function testRuleViolationDoesNotSave(): void
    {
        $id = $this->service->create('Ana');

        try {
            $this->service->place($id);
            self::fail('Esperava uma violação de regra');
        } catch (BusinessRuleViolation) {
            self::assertSame(OrderStatus::Draft, $this->repository->saved[$id]->status());
            self::assertSame([], $this->events->published);
        }
    }

    public function testUnknownOrderIsNotFound(): void
    {
        $this->expectException(OrderNotFound::class);

        $this->service->place(Identifier::generate());
    }

    public function testInvalidIdIsARuleViolation(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('Identificador inválido.'));

        $this->service->get('abc');
    }

    public function testListLimitIsBounded(): void
    {
        $this->service->create('Ana');
        $this->service->listRecent(0);
        $this->service->listRecent(5000);
        $this->service->listRecent(10);

        self::assertSame([ListLimit::DEFAULT, ListLimit::MAX, 10], $this->repository->limits);
    }
}
