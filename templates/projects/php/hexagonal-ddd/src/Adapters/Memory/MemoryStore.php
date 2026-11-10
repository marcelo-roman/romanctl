<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Memory;

final readonly class MemoryStore
{
    public function __construct(private ?string $directory = null)
    {
    }

    public function load(string $collection): array
    {
        $path = $this->path($collection);
        if ($path === null || !is_file($path)) {
            return [];
        }
        $items = unserialize((string) file_get_contents($path));

        return is_array($items) ? $items : [];
    }

    public function save(string $collection, array $items): void
    {
        $path = $this->path($collection);
        if ($path === null) {
            return;
        }
        if (!is_dir(dirname($path))) {
            mkdir(dirname($path), 0o777, true);
        }
        file_put_contents($path, serialize($items), LOCK_EX);
    }

    private function path(string $collection): ?string
    {
        if ($this->directory === null) {
            return null;
        }

        return "{$this->directory}/{$collection}.data";
    }
}
