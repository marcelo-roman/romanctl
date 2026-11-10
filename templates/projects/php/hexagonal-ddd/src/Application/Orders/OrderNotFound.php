<?php

declare(strict_types=1);

namespace SampleApp\Application\Orders;

use SampleApp\Domain\Shared\NotFound;

final class OrderNotFound extends NotFound
{
    public function __construct()
    {
        parent::__construct('Pedido não encontrado.');
    }
}
