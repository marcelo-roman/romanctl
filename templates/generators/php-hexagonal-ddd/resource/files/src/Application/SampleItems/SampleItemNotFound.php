<?php

declare(strict_types=1);

namespace SampleApp\Application\SampleItems;

use SampleApp\Domain\Shared\NotFound;

final class SampleItemNotFound extends NotFound
{
    public function __construct()
    {
        parent::__construct('SampleItem não encontrado.');
    }
}
