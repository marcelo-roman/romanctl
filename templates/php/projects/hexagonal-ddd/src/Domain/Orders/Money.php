<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

use SampleApp\Domain\Shared\BusinessRuleViolation;

final readonly class Money
{
    public const string DEFAULT_CURRENCY = 'BRL';

    public int $cents;
    public string $currency;

    public function __construct(int $cents, string $currency = self::DEFAULT_CURRENCY)
    {
        if ($cents < 0) {
            throw new BusinessRuleViolation('O valor não pode ser negativo.');
        }
        $normalized = strtoupper(trim($currency));
        if (preg_match('/^[A-Z]{3}$/', $normalized) !== 1) {
            throw new BusinessRuleViolation('A moeda deve ter 3 letras, como BRL.');
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
            throw new BusinessRuleViolation('Não é possível somar moedas diferentes.');
        }

        return new self($this->cents + $other->cents, $this->currency);
    }

    public function equals(self $other): bool
    {
        return $this->cents === $other->cents && $this->currency === $other->currency;
    }
}
