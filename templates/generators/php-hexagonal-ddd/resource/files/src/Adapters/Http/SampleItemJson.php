<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use DateTimeInterface;
use SampleApp\Application\SampleItems\SampleItemView;

final class SampleItemJson
{
    private function __construct()
    {
    }

    public static function of(SampleItemView $view): array
    {
        return [
            'id' => $view->id,
            'name' => $view->name,
            'createdAt' => $view->createdAt->format(DateTimeInterface::ATOM),
        ];
    }
}
