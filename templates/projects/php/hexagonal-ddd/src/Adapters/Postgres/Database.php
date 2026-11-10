<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Postgres;

use Doctrine\DBAL\Connection;
use Doctrine\DBAL\DriverManager;
use Doctrine\DBAL\Tools\DsnParser;

final class Database
{
    private const int SCHEMA_LOCK = 72_616_401;
    private const array SCHEMES = ['postgres' => 'pdo_pgsql', 'postgresql' => 'pdo_pgsql', 'pgsql' => 'pdo_pgsql'];

    private ?Connection $connection = null;

    public function __construct(private readonly string $url)
    {
    }

    public function connection(): Connection
    {
        if ($this->connection !== null) {
            return $this->connection;
        }
        $connection = DriverManager::getConnection((new DsnParser(self::SCHEMES))->parse($this->url));
        self::createSchema($connection);

        return $this->connection = $connection;
    }

    private static function createSchema(Connection $connection): void
    {
        $schema = (string) file_get_contents(__DIR__ . '/schema.sql');
        $connection->transactional(static function (Connection $transaction) use ($schema): void {
            $transaction->executeQuery('SELECT pg_advisory_xact_lock(?)', [self::SCHEMA_LOCK]);
            $transaction->executeStatement($schema);
        });
    }
}
