<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use Psr\Log\LoggerInterface;
use SampleApp\Domain\Shared\BusinessRuleViolation;
use SampleApp\Domain\Shared\NotFound;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Event\ExceptionEvent;
use Symfony\Component\HttpKernel\Exception\HttpExceptionInterface;
use Symfony\Component\HttpKernel\KernelEvents;
use Throwable;

final readonly class ProblemSubscriber implements EventSubscriberInterface
{
    public function __construct(private LoggerInterface $logger)
    {
    }

    public static function getSubscribedEvents(): array
    {
        return [KernelEvents::EXCEPTION => 'onException'];
    }

    public function onException(ExceptionEvent $event): void
    {
        $event->setResponse($this->problemFor($event->getThrowable()));
    }

    private function problemFor(Throwable $error): Response
    {
        if ($error instanceof InvalidRequest) {
            return ApiResponse::problem(Response::HTTP_BAD_REQUEST, 'Requisição inválida', $error->getMessage());
        }
        if ($error instanceof BusinessRuleViolation) {
            return ApiResponse::problem(Response::HTTP_UNPROCESSABLE_ENTITY, 'Regra de negócio violada', $error->getMessage());
        }
        if ($error instanceof NotFound) {
            return ApiResponse::problem(Response::HTTP_NOT_FOUND, 'Não encontrado', $error->getMessage());
        }
        if ($error instanceof HttpExceptionInterface) {
            return ApiResponse::problem($error->getStatusCode(), 'Requisição não atendida', $error->getMessage());
        }
        $this->logger->error('request failed: {message}', ['message' => $error->getMessage(), 'exception' => $error]);

        return ApiResponse::problem(Response::HTTP_INTERNAL_SERVER_ERROR, 'Erro interno');
    }
}
