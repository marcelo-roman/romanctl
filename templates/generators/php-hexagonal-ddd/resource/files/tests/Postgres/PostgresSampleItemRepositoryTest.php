<?php

declare(strict_types=1);

namespace SampleApp\Tests\Postgres;

use DateTimeImmutable;
use DateTimeZone;
use SampleApp\Adapters\Postgres\PostgresSampleItemRepository;
use SampleApp\Domain\SampleItems\SampleItem;
use SampleApp\Domain\Shared\Identifier;

final class PostgresSampleItemRepositoryTest extends DatabaseTestCase
{
    public function testSampleItemRoundTrip(): void
    {
        $repository = new PostgresSampleItemRepository($this->database);
        $sampleItem = SampleItem::create('Primeiro', new DateTimeImmutable('now', new DateTimeZone('UTC')));
        $repository->save($sampleItem);
        $sampleItem->rename('Segundo');
        $repository->save($sampleItem);

        $loaded = $repository->find($sampleItem->id());

        self::assertNotNull($loaded);
        self::assertSame('Segundo', $loaded->name());
        self::assertEquals($sampleItem->createdAt(), $loaded->createdAt());
        self::assertContains($sampleItem->id(), array_map(static fn (SampleItem $recent) => $recent->id(), $repository->listRecent(200)));
        self::assertTrue($repository->delete($sampleItem->id()));
        self::assertFalse($repository->delete($sampleItem->id()));
        self::assertNull($repository->find(Identifier::generate()));
    }
}
