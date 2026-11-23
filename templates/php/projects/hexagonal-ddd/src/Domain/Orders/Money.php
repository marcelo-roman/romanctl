<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

use SampleApp\Domain\Shared\BusinessRuleViolation;

final readonly class Money
{
    public const string DEFAULT_CURRENCY = 'USD';

    public int $cents;
    public string $currency;

    public function __construct(int $cents, string $currency = self::DEFAULT_CURRENCY)
    {
        if ($cents < 0) {
            throw new BusinessRuleViolation('Amount cannot be negative.');
        }
        $normalized = strtoupper(trim($currency));
        if (preg_match('/^[A-Z]{3}$/', $normalized) !== 1) {
            throw new BusinessRuleViolation('Currency must have 3 letters, like USD.');
        }
        $this->cents = $cents;
        $this->currency = $normalized;
    }

    public static function fromAmount(float|int $amount, string $currency = ''): self
    {
        return new self((int) round($amount * 100), $currency === '' ? self::DEFAULT_CURRENCY : $currency);
    }

    public static function zero(string $currency = self::DEFAULT_CURRENCY): self
    {
        return new self(0, $currency);
    }

    public function amount(): float
    {
        return $this->cents / 100;
    }

    public function multiply(int $quantity): self
    {
        return new self($this->cents * $quantity, $this->currency);
    }

    public function add(self $other): self
    {
        if ($other->currency !== $this->currency) {
            throw new BusinessRuleViolation('Cannot add different currencies.');
        }

        return new self($this->cents + $other->cents, $this->currency);
    }

    public function equals(self $other): bool
    {
        return $this->cents === $other->cents && $this->currency === $other->currency;
    }
}
