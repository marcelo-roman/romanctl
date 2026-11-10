<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use PHPUnit\Framework\TestCase;
use SampleApp\Application\SampleItems\SampleItemNotFound;
use SampleApp\Application\SampleItems\SampleItemService;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final class SampleItemServiceTest extends TestCase
{
    private SampleItemService $service;

    protected function setUp(): void
    {
        $this->service = new SampleItemService(new FakeSampleItemRepository(), new FixedClock());
    }

    public function testCreateRenameAndDelete(): void
    {
        $id = $this->service->create('Primeiro');
        $this->service->rename($id, 'Segundo');

        self::assertSame('Segundo', $this->service->get($id)->name);
        self::assertSame([$id], array_map(static fn ($view) => $view->id, $this->service->listRecent()));

        $this->service->delete($id);
        $this->expectException(SampleItemNotFound::class);

        $this->service->get($id);
    }

    public function testUnknownSampleItemIsNotFound(): void
    {
        $this->expectException(SampleItemNotFound::class);

        $this->service->delete(Identifier::generate());
    }

    public function testInvalidIdIsARuleViolation(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('Identificador inválido.'));

        $this->service->rename('abc', 'Nome');
    }
}
