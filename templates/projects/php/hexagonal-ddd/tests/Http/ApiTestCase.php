<?php

declare(strict_types=1);

namespace SampleApp\Tests\Http;

use Symfony\Bundle\FrameworkBundle\KernelBrowser;
use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;
use Symfony\Component\HttpFoundation\Response;

abstract class ApiTestCase extends WebTestCase
{
    protected KernelBrowser $client;

    protected function setUp(): void
    {
        $this->client = static::createClient();
        $this->client->disableReboot();
    }

    protected function send(string $method, string $uri, mixed $body = null): Response
    {
        $content = is_string($body) ? $body : json_encode($body, JSON_THROW_ON_ERROR);
        $this->client->request($method, $uri, server: ['CONTENT_TYPE' => 'application/json'], content: $body === null ? null : $content);

        return $this->client->getResponse();
    }

    protected function json(Response $response): mixed
    {
        return json_decode((string) $response->getContent(), true, 64, JSON_THROW_ON_ERROR);
    }
}
