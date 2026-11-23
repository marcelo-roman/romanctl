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
        self::assertSame('First', SampleItem::create('  First ', new DateTimeImmutable())->name());
    }

    public function testCreateRequiresName(): void
    {
        $this->expectExceptionObject(new BusinessRuleViolation('Name is required.'));

        SampleItem::create(' ', new DateTimeImmutable());
    }

    public function testRenameRequiresName(): void
    {
        $sampleItem = SampleItem::create('First', new DateTimeImmutable());
        $sampleItem->rename('Second');

        self::assertSame('Second', $sampleItem->name());
        $this->expectException(BusinessRuleViolation::class);

        $sampleItem->rename('');
    }
}
