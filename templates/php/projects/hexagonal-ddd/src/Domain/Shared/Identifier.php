<?php

declare(strict_types=1);

namespace SampleApp\Domain\Shared;

final class Identifier
{
    private const string PATTERN = '/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/';

    private function __construct()
    {
    }

    public static function generate(): string
    {
        $bytes = random_bytes(16);
        $bytes[6] = chr((ord($bytes[6]) & 0x0f) | 0x40);
        $bytes[8] = chr((ord($bytes[8]) & 0x3f) | 0x80);

        return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($bytes), 4));
    }

    public static function parse(string $value): string
    {
        $normalized = strtolower(trim($value));
        if (preg_match(self::PATTERN, $normalized) !== 1) {
            throw new BusinessRuleViolation('Identificador inválido.');
        }

        return $normalized;
    }
}
