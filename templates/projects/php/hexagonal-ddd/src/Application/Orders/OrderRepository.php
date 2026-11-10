<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

use SampleApp\Domain\Orders\Order;

interface OrderRepository
{
    public function save(Order $order): void;

    public function find(string $id): ?Order;

    public function listRecent(int $limit): array;
}
