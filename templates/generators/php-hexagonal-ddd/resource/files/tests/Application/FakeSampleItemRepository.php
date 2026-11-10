<?php

declare(strict_types=1);

namespace SampleApp\Tests\Application;

use SampleApp\Application\SampleItems\SampleItemRepository;
use SampleApp\Domain\SampleItems\SampleItem;

final class FakeSampleItemRepository implements SampleItemRepository
{
    public array $saved = [];

    public function save(SampleItem $sampleItem): void
    {
        $this->saved[$sampleItem->id()] = $sampleItem;
    }

    public function find(string $id): ?SampleItem
    {
        return $this->saved[$id] ?? null;
    }

    public function listRecent(int $limit): array
    {
        return array_slice(array_values($this->saved), 0, $limit);
    }

    public function delete(string $id): bool
    {
        if (!isset($this->saved[$id])) {
            return false;
        }
        unset($this->saved[$id]);

        return true;
    }
}
