<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use DateTimeInterface;
use SampleApp\Application\Orders\OrderItemView;
use SampleApp\Application\Orders\OrderView;

final class OrderJson
{
    private function __construct()
    {
    }

    public static function of(OrderView $view): array
    {
        return [
            'id' => $view->id,
            'customerName' => $view->customerName,
            'status' => $view->status,
            'createdAt' => $view->createdAt->format(DateTimeInterface::ATOM),
            'total' => $view->total,
            'currency' => $view->currency,
            'items' => array_map(self::item(...), $view->items),
        ];
    }

    private static function item(OrderItemView $item): array
    {
        return [
            'productName' => $item->productName,
            'quantity' => $item->quantity,
            'unitPrice' => $item->unitPrice,
            'subtotal' => $item->subtotal,
        ];
    }
}
