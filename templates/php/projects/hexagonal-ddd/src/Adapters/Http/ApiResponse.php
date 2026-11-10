<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;

final class ApiResponse
{
    private const int ENCODING = JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_PRESERVE_ZERO_FRACTION;

    private function __construct()
    {
    }

    public static function json(mixed $data, int $status = Response::HTTP_OK, array $headers = []): JsonResponse
    {
        $response = new JsonResponse(null, $status, $headers);
        $response->setEncodingOptions(self::ENCODING);

        return $response->setData($data);
    }

    public static function created(string $id, string $location): JsonResponse
    {
        return self::json(['id' => $id], Response::HTTP_CREATED, ['Location' => $location]);
    }

    public static function noContent(): Response
    {
        return new Response(null, Response::HTTP_NO_CONTENT);
    }

    public static function problem(int $status, string $title, string $detail = ''): JsonResponse
    {
        $response = self::json(['status' => $status, 'title' => $title, 'detail' => $detail], $status);
        $response->headers->set('Content-Type', 'application/problem+json');

        return $response;
    }
}
