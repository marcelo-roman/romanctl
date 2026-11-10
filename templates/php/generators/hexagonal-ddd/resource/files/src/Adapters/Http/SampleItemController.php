<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use SampleApp\Application\SampleItems\SampleItemService;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;

#[AsController]
#[Route('/sample-items')]
final readonly class SampleItemController
{
    public function __construct(private SampleItemService $sampleItems)
    {
    }

    #[Route('', methods: ['POST'])]
    public function create(Request $request): JsonResponse
    {
        $id = $this->sampleItems->create(JsonBody::of($request)->string('name'));

        return ApiResponse::created($id, "/sample-items/{$id}");
    }

    #[Route('', methods: ['GET'])]
    public function list(Request $request): JsonResponse
    {
        $views = $this->sampleItems->listRecent($request->query->getInt('limit'));

        return ApiResponse::json(array_map(SampleItemJson::of(...), $views));
    }

    #[Route('/{id}', methods: ['GET'])]
    public function get(string $id): JsonResponse
    {
        return ApiResponse::json(SampleItemJson::of($this->sampleItems->get($id)));
    }

    #[Route('/{id}', methods: ['PUT'])]
    public function rename(string $id, Request $request): Response
    {
        $this->sampleItems->rename($id, JsonBody::of($request)->string('name'));

        return ApiResponse::noContent();
    }

    #[Route('/{id}', methods: ['DELETE'])]
    public function delete(string $id): Response
    {
        $this->sampleItems->delete($id);

        return ApiResponse::noContent();
    }
}
