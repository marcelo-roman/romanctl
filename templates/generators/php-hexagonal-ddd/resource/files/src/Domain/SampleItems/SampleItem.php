<?php

declare(strict_types=1);

namespace SampleApp\Domain\SampleItems;

use DateTimeImmutable;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final class SampleItem
{
    private function __construct(
        private readonly string $id,
        private string $name,
        private readonly DateTimeImmutable $createdAt,
    ) {
    }

    public static function create(string $name, DateTimeImmutable $now): self
    {
        return new self(Identifier::generate(), self::validName($name), $now);
    }

    public static function restore(string $id, string $name, DateTimeImmutable $createdAt): self
    {
        return new self($id, $name, $createdAt);
    }

    public function id(): string
    {
        return $this->id;
    }

    public function name(): string
    {
        return $this->name;
    }

    public function createdAt(): DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function rename(string $name): void
    {
        $this->name = self::validName($name);
    }

    private static function validName(string $name): string
    {
        if (trim($name) === '') {
            throw new BusinessRuleViolation('O nome é obrigatório.');
        }

        return trim($name);
    }
}
