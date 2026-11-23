<?php

declare(strict_types=1);

namespace SampleApp\Adapters\Http;

use JsonException;
use Symfony\Component\HttpFoundation\Request;

final readonly class JsonBody
{
    private function __construct(private array $fields)
    {
    }

    public static function of(Request $request): self
    {
        try {
            $decoded = json_decode($request->getContent(), true, 64, JSON_THROW_ON_ERROR);
        } catch (JsonException $error) {
            throw new InvalidRequest('Invalid JSON: ' . $error->getMessage());
        }
        if (!is_array($decoded) || ($decoded !== [] && array_is_list($decoded))) {
            throw new InvalidRequest('The body must be a JSON object.');
        }

        return new self($decoded);
    }

    public function string(string $name): string
    {
        $value = $this->fields[$name] ?? null;
        if (!is_string($value)) {
            throw new InvalidRequest("{$name}: must be a string.");
        }

        return $value;
    }

    public function optionalString(string $name, string $default = ''): string
    {
        if (!isset($this->fields[$name])) {
            return $default;
        }

        return $this->string($name);
    }

    public function integer(string $name): int
    {
        $value = $this->fields[$name] ?? null;
        if (!is_int($value)) {
            throw new InvalidRequest("{$name}: must be an integer.");
        }

        return $value;
    }

    public function number(string $name): float|int
    {
        $value = $this->fields[$name] ?? null;
        if (!is_int($value) && !is_float($value)) {
            throw new InvalidRequest("{$name}: must be a number.");
        }

        return $value;
    }
}
