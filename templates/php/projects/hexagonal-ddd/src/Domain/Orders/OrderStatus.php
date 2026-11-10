<?php

declare(strict_types=1);

namespace SampleApp\Domain\Orders;

enum OrderStatus: string
{
    case Draft = 'Draft';
    case Placed = 'Placed';
}
