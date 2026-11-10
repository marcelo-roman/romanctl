<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Postgres;

use DateTimeImmutable;
use DateTimeZone;
use Doctrine\DBAL\ParameterType;
use SampleApp\Application\SampleItems\SampleItemRepository;
use SampleApp\Domain\SampleItems\SampleItem;

final readonly class PostgresSampleItemRepository implements SampleItemRepository
{
    private const string UPSERT = <<<'SQL'
        INSERT INTO sample_items (id, name, created_at) VALUES (?, ?, ?)
        ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
        SQL;
    private const string SELECT = 'SELECT id, name, created_at FROM sample_items';

    public function __construct(private Database $database)
    {
    }

    public function save(SampleItem $sampleItem): void
    {
        $this->database->connection()->executeStatement(self::UPSERT, [
            $sampleItem->id(),
            $sampleItem->name(),
            $sampleItem->createdAt()->format('Y-m-d H:i:s.uP'),
        ]);
    }

    public function find(string $id): ?SampleItem
    {
        return $this->query(self::SELECT . ' WHERE id = ?', [$id], [ParameterType::STRING])[0] ?? null;
    }

    public function listRecent(int $limit): array
    {
        return $this->query(self::SELECT . ' ORDER BY created_at DESC LIMIT ?', [$limit], [ParameterType::INTEGER]);
    }

    public function delete(string $id): bool
    {
        return $this->database->connection()->executeStatement('DELETE FROM sample_items WHERE id = ?', [$id]) > 0;
    }

    private function query(string $query, array $params, array $types): array
    {
        return array_map(
            static fn (array $row) => SampleItem::restore(
                $row['id'],
                $row['name'],
                (new DateTimeImmutable($row['created_at']))->setTimezone(new DateTimeZone('UTC')),
            ),
            $this->database->connection()->fetchAllAssociative($query, $params, $types),
        );
    }
}
