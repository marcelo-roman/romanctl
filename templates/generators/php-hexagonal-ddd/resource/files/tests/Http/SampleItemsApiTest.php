<?php

declare(strict_types=1);

namespace SampleApp\Tests\Http;

final class SampleItemsApiTest extends ApiTestCase
{
    public function testSampleItemCrud(): void
    {
        $response = $this->send('POST', '/sample-items', ['name' => 'Primeiro']);
        self::assertSame(201, $response->getStatusCode());
        $id = $this->json($response)['id'];
        self::assertSame("/sample-items/{$id}", $response->headers->get('Location'));

        self::assertSame(204, $this->send('PUT', "/sample-items/{$id}", ['name' => 'Segundo'])->getStatusCode());
        self::assertSame('Segundo', $this->json($this->send('GET', "/sample-items/{$id}"))['name']);
        self::assertSame([$id], array_column($this->json($this->send('GET', '/sample-items')), 'id'));

        self::assertSame(204, $this->send('DELETE', "/sample-items/{$id}")->getStatusCode());
        self::assertSame(404, $this->send('GET', "/sample-items/{$id}")->getStatusCode());
    }

    public function testSampleItemErrors(): void
    {
        self::assertSame(422, $this->send('POST', '/sample-items', ['name' => ' '])->getStatusCode());
        self::assertSame(400, $this->send('POST', '/sample-items', '{}')->getStatusCode());
        self::assertSame(422, $this->send('GET', '/sample-items/abc')->getStatusCode());
        self::assertSame(404, $this->send('DELETE', '/sample-items/00000000-0000-4000-8000-000000000000')->getStatusCode());
    }
}
