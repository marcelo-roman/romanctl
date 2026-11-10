<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Memory;

use SampleApp\Application\SampleItems\SampleItemRepository;
use SampleApp\Domain\SampleItems\SampleItem;

final class InMemorySampleItemRepository implements SampleItemRepository
{
    private const string COLLECTION = 'sample_items';

    private array $sampleItems;

    public function __construct(private readonly MemoryStore $store = new MemoryStore())
    {
        $this->sampleItems = $store->load(self::COLLECTION);
    }

    public function save(SampleItem $sampleItem): void
    {
        $this->sampleItems[$sampleItem->id()] = self::copy($sampleItem);
        $this->store->save(self::COLLECTION, $this->sampleItems);
    }

    public function find(string $id): ?SampleItem
    {
        if (!isset($this->sampleItems[$id])) {
            return null;
        }

        return self::copy($this->sampleItems[$id]);
    }

    public function listRecent(int $limit): array
    {
        $sampleItems = array_reverse(array_values($this->sampleItems));
        usort($sampleItems, static fn (SampleItem $a, SampleItem $b) => $b->createdAt() <=> $a->createdAt());

        return array_map(self::copy(...), array_slice($sampleItems, 0, $limit));
    }

    public function delete(string $id): bool
    {
        if (!isset($this->sampleItems[$id])) {
            return false;
        }
        unset($this->sampleItems[$id]);
        $this->store->save(self::COLLECTION, $this->sampleItems);

        return true;
    }

    private static function copy(SampleItem $sampleItem): SampleItem
    {
        return SampleItem::restore($sampleItem->id(), $sampleItem->name(), $sampleItem->createdAt());
    }
}
