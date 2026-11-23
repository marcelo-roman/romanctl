<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\Identifier;

final readonly class OrderItem
{
    public function __construct(
        public string $id,
        public string $productName,
        public int $quantity,
        public Money $unitPrice,
    ) {
    }

    public static function create(string $productName, int $quantity, Money $unitPrice): self
    {
        if (trim($productName) === '') {
            throw new BusinessRuleViolation('Product name is required.');
        }
        if ($quantity <= 0) {
            throw new BusinessRuleViolation('Quantity must be greater than zero.');
        }

        return new self(Identifier::generate(), trim($productName), $quantity, $unitPrice);
    }

    public function subtotal(): Money
    {
        return $this->unitPrice->multiply($this->quantity);
    }
}
