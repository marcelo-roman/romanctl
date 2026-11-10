<?php

declare(strict_types=1);

namespace SampleApp\Application\SampleItems;

use SampleApp\Application\Shared\Clock;
use SampleApp\Application\Shared\ListLimit;
use SampleApp\Domain\SampleItems\SampleItem;
use SampleApp\Domain\Shared\Identifier;

final readonly class SampleItemService
{
    public function __construct(
        private SampleItemRepository $repository,
        private Clock $clock,
    ) {
    }

    public function create(string $name): string
    {
        $sampleItem = SampleItem::create($name, $this->clock->now());
        $this->repository->save($sampleItem);

        return $sampleItem->id();
    }

    public function rename(string $id, string $name): void
    {
        $sampleItem = $this->load($id);
        $sampleItem->rename($name);
        $this->repository->save($sampleItem);
    }

    public function get(string $id): SampleItemView
    {
        return SampleItemView::of($this->load($id));
    }

    public function listRecent(int $limit = ListLimit::DEFAULT): array
    {
        return array_map(SampleItemView::of(...), $this->repository->listRecent(ListLimit::bound($limit)));
    }

    public function delete(string $id): void
    {
        if (!$this->repository->delete(Identifier::parse($id))) {
            throw new SampleItemNotFound();
        }
    }

    private function load(string $id): SampleItem
    {
        $sampleItem = $this->repository->find(Identifier::parse($id));
        if ($sampleItem === null) {
            throw new SampleItemNotFound();
        }

        return $sampleItem;
    }
}
