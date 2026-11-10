<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use SampleApp\Application\Orders\AddItem;
use SampleApp\Application\Orders\OrderService;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;

#[AsController]
#[Route('/orders')]
final readonly class OrderController
{
    public function __construct(private OrderService $orders)
    {
    }

    #[Route('', methods: ['POST'])]
    public function create(Request $request): JsonResponse
    {
        $id = $this->orders->create(JsonBody::of($request)->string('customerName'));

        return ApiResponse::created($id, "/orders/{$id}");
    }

    #[Route('/{id}/items', methods: ['POST'])]
    public function addItem(string $id, Request $request): Response
    {
        $body = JsonBody::of($request);
        $command = new AddItem(
            $body->string('productName'),
            $body->integer('quantity'),
            $body->number('unitPrice'),
            $body->optionalString('currency'),
        );
        $this->orders->addItem($id, $command);

        return ApiResponse::noContent();
    }

    #[Route('/{id}/place', methods: ['POST'])]
    public function place(string $id): Response
    {
        $this->orders->place($id);

        return ApiResponse::noContent();
    }

    #[Route('/{id}', methods: ['GET'])]
    public function get(string $id): JsonResponse
    {
        return ApiResponse::json(OrderJson::of($this->orders->get($id)));
    }

    #[Route('', methods: ['GET'])]
    public function list(Request $request): JsonResponse
    {
        $views = $this->orders->listRecent($request->query->getInt('limit'));

        return ApiResponse::json(array_map(OrderJson::of(...), $views));
    }
}
