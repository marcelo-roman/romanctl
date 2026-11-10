<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Postgres;

use DateTimeImmutable;
use DateTimeZone;
use Doctrine\DBAL\ArrayParameterType;
use Doctrine\DBAL\Connection;
use Doctrine\DBAL\ParameterType;
use SampleApp\Application\Orders\OrderRepository;
use SampleApp\Domain\Orders\Money;
use SampleApp\Domain\Orders\Order;
use SampleApp\Domain\Orders\OrderItem;
use SampleApp\Domain\Orders\OrderStatus;

final readonly class PostgresOrderRepository implements OrderRepository
{
    private const string UPSERT_ORDER = <<<'SQL'
        INSERT INTO orders (id, customer_name, status, created_at) VALUES (?, ?, ?, ?)
        ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status
        SQL;
    private const string INSERT_ITEM = <<<'SQL'
        INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        SQL;
    private const string SELECT_ORDERS = 'SELECT id, customer_name, status, created_at FROM orders';
    private const string SELECT_ITEMS = <<<'SQL'
        SELECT order_id, id, product_name, quantity, unit_price, currency
        FROM order_items WHERE order_id IN (?) ORDER BY position
        SQL;
    private const string TIMESTAMP = 'Y-m-d H:i:s.uP';

    public function __construct(private Database $database)
    {
    }

    public function save(Order $order): void
    {
        $this->database->connection()->transactional(static function (Connection $connection) use ($order): void {
            $connection->executeStatement(self::UPSERT_ORDER, [
                $order->id(),
                $order->customerName(),
                $order->status()->value,
                $order->createdAt()->format(self::TIMESTAMP),
            ]);
            $connection->executeStatement('DELETE FROM order_items WHERE order_id = ?', [$order->id()]);
            foreach ($order->items() as $position => $item) {
                $connection->executeStatement(self::INSERT_ITEM, [
                    $item->id,
                    $order->id(),
                    $position,
                    $item->productName,
                    $item->quantity,
                    $item->unitPrice->cents,
                    $item->unitPrice->currency,
                ]);
            }
        });
    }

    public function find(string $id): ?Order
    {
        return $this->load(self::SELECT_ORDERS . ' WHERE id = ?', [$id], [ParameterType::STRING])[0] ?? null;
    }

    public function listRecent(int $limit): array
    {
        return $this->load(self::SELECT_ORDERS . ' ORDER BY created_at DESC LIMIT ?', [$limit], [ParameterType::INTEGER]);
    }

    private function load(string $query, array $params, array $types): array
    {
        $connection = $this->database->connection();
        $heads = $connection->fetchAllAssociative($query, $params, $types);
        $items = self::items($connection, array_column($heads, 'id'));

        return array_map(
            static fn (array $head) => Order::restore(
                $head['id'],
                $head['customer_name'],
                OrderStatus::from($head['status']),
                self::timestamp($head['created_at']),
                ...$items[$head['id']] ?? [],
            ),
            $heads,
        );
    }

    private static function items(Connection $connection, array $orderIds): array
    {
        if ($orderIds === []) {
            return [];
        }
        $items = [];
        $rows = $connection->fetchAllAssociative(self::SELECT_ITEMS, [$orderIds], [ArrayParameterType::STRING]);
        foreach ($rows as $row) {
            $items[$row['order_id']][] = new OrderItem(
                $row['id'],
                $row['product_name'],
                (int) $row['quantity'],
                new Money((int) $row['unit_price'], $row['currency']),
            );
        }

        return $items;
    }

    private static function timestamp(string $value): DateTimeImmutable
    {
        return (new DateTimeImmutable($value))->setTimezone(new DateTimeZone('UTC'));
    }
}
