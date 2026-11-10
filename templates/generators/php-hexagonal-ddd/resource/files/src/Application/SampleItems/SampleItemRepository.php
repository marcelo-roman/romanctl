<?php

declare(strict_types=1);

namespace SampleApp\Application\SampleItems;

use SampleApp\Domain\SampleItems\SampleItem;

interface SampleItemRepository
{
    public function save(SampleItem $sampleItem): void;

    public function find(string $id): ?SampleItem;

    public function listRecent(int $limit): array;

    public function delete(string $id): bool;
}
