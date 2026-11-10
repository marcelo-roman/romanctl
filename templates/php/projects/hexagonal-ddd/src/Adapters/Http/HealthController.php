<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;

#[AsController]
final class HealthController
{
    #[Route('/health', methods: ['GET'])]
    public function __invoke(): JsonResponse
    {
        return ApiResponse::json(['status' => 'ok']);
    }
}
