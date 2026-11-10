<?php

declare(strict_types=1);

namespace SampleApp\Application\Shared;

final class ListLimit
{
    public const int DEFAULT = 50;
    public const int MAX = 200;

    private function __construct()
    {
    }

    public static function bound(int $requested): int
    {
        if ($requested <= 0) {
            return self::DEFAULT;
        }

        return min($requested, self::MAX);
    }
}
