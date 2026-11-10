<?php

declare(strict_types=1);

namespace SampleApp\Tests\Postgres;

use PHPUnit\Framework\TestCase;
use SampleApp\Adapters\Postgres\Database;

abstract class DatabaseTestCase extends TestCase
{
    protected Database $database;

    protected function setUp(): void
    {
        $url = getenv('DATABASE_URL');
        if ($url === false || $url === '') {
            self::markTestSkipped('DATABASE_URL not set: romanctl up sample-app --no-run creates the database and the .env');
        }
        $this->database = new Database($url);
    }

    protected function newDatabase(): Database
    {
        return new Database((string) getenv('DATABASE_URL'));
    }
}
