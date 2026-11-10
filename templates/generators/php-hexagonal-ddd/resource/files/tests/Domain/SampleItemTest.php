<?php

declare(strict_types=1);

namespace SampleApp\Tests\Domain;

use DateTimeImmutable;
use PHPUnit\Framework\TestCase;
use SampleApp\Domain\SampleItems\SampleItem;
use SampleApp\Domain\Shared\BusinessRuleViolation;

final class SampleItemTest extends TestCase
{
    public function testCreateTrimsName(): void
    {
        self::assertSame('Primeiro', SampleItem::create('  Primeiro ', new DateTimeImmutable())->name());
    }

    public function testCreateRequiresName(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('O nome é obrigatório.'));

        SampleItem::create(' ', new DateTimeImmutable());
    }

    public function testRenameRequiresName(): void
    {
        $sampleItem = SampleItem::create('Primeiro', new DateTimeImmutable());
        $sampleItem->rename('Segundo');

        self::assertSame('Segundo', $sampleItem->name());
        $this->expectException(BusinessRuleViolation::class);

        $sampleItem->rename('');
    }
}
