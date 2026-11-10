<?php

declare(strict_types=1);

namespace SampleApp\Application\SampleItems;

use DateTimeImmutable;
use SampleApp\Domain\SampleItems\SampleItem;

final readonly class SampleItemView
{
    public function __construct(
        public string $id,
        public string $name,
        public DateTimeImmutable $createdAt,
    ) {
    }

    public static function of(SampleItem $sampleItem): self
    {
        return new self($sampleItem->id(), $sampleItem->name(), $sampleItem->createdAt());
    }
}
