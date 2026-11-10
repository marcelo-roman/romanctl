<?php

declare(strict_types=1);

namespace SampleApp\Tests\Http;

use PHPUnit\Framework\Attributes\DataProvider;

final class OrdersApiTest extends ApiTestCase
{
    private const string MISSING_ID = '00000000-0000-4000-8000-000000000000';

    public function testHealth(): void
    {
        $response = $this->send('GET', '/health');

        self::assertSame(200, $response->getStatusCode());
        self::assertSame(['status' => 'ok'], $this->json($response));
    }

    public function testOrderLifecycle(): void
    {
        $response = $this->send('POST', '/orders', ['customerName' => 'Ana']);
        self::assertSame(201, $response->getStatusCode());
        $id = $this->json($response)['id'];
        self::assertSame("/orders/{$id}", $response->headers->get('Location'));

        $item = ['productName' => 'Caneta', 'quantity' => 2, 'unitPrice' => 3.5];
        self::assertSame(204, $this->send('POST', "/orders/{$id}/items", $item)->getStatusCode());
        self::assertSame(204, $this->send('POST', "/orders/{$id}/place")->getStatusCode());

        $order = $this->json($this->send('GET', "/orders/{$id}"));
        self::assertSame('Placed', $order['status']);
        self::assertSame('Ana', $order['customerName']);
        self::assertEquals(7, $order['total']);
        self::assertSame('BRL', $order['currency']);
        self::assertEquals([['productName' => 'Caneta', 'quantity' => 2, 'unitPrice' => 3.5, 'subtotal' => 7]], $order['items']);
    }

    public function testListReturnsMostRecentFirst(): void
    {
        $first = $this->createOrder('Ana');
        $second = $this->createOrder('Bia');

        self::assertSame([$second, $first], array_column($this->json($this->send('GET', '/orders')), 'id'));
        self::assertSame([$second], array_column($this->json($this->send('GET', '/orders?limit=1')), 'id'));
    }

    public function testBusinessRuleViolationIs422(): void
    {
        $response = $this->send('POST', '/orders', ['customerName' => '']);

        self::assertSame(422, $response->getStatusCode());
        self::assertSame('application/problem+json', $response->headers->get('Content-Type'));
        self::assertSame(
            ['status' => 422, 'title' => 'Regra de negócio violada', 'detail' => 'O nome do cliente é obrigatório.'],
            $this->json($response),
        );
    }

    public function testPlacingAnEmptyOrderIs422(): void
    {
        $id = $this->createOrder();

        self::assertSame(422, $this->send('POST', "/orders/{$id}/place")->getStatusCode());
    }

    public function testMalformedIdIs422(): void
    {
        $response = $this->send('GET', '/orders/abc');

        self::assertSame(422, $response->getStatusCode());
        self::assertSame('Identificador inválido.', $this->json($response)['detail']);
    }

    public function testMissingOrderIs404(): void
    {
        $response = $this->send('GET', '/orders/' . self::MISSING_ID);

        self::assertSame(404, $response->getStatusCode());
        self::assertSame('Pedido não encontrado.', $this->json($response)['detail']);
    }

    #[DataProvider('invalidBodies')]
    public function testInvalidBodyIs400(string $body): void
    {
        $response = $this->send('POST', '/orders', $body);

        self::assertSame(400, $response->getStatusCode());
        self::assertSame('Requisição inválida', $this->json($response)['title']);
    }

    public static function invalidBodies(): iterable
    {
        yield 'malformed' => ['{'];
        yield 'missing field' => ['{"quantity": 1}'];
        yield 'null field' => ['{"customerName": null}'];
        yield 'list' => ['[1]'];
    }

    public function testInvalidItemIs400(): void
    {
        $id = $this->createOrder();
        $item = ['productName' => 'Caneta', 'quantity' => 'dois', 'unitPrice' => 3.5];

        self::assertSame(400, $this->send('POST', "/orders/{$id}/items", $item)->getStatusCode());
    }

    private function createOrder(string $customer = 'Ana'): string
    {
        $response = $this->send('POST', '/orders', ['customerName' => $customer]);
        self::assertSame(201, $response->getStatusCode());

        return $this->json($response)['id'];
    }
}
