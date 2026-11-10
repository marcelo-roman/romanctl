<?php

declare(strict_types=1);

namespace SampleApp\Tests\Architecture;

use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use SplFileInfo;

final readonly class LayerRules
{
    private const array ALLOWED = [
        'Domain' => ['Domain'],
        'Application' => ['Domain', 'Application'],
        'Adapters' => ['Domain', 'Application', 'Adapters'],
    ];
    private const array OWN_CODE_ONLY = ['Domain', 'Application'];

    public function __construct(private string $root)
    {
    }

    public function violations(string $source): array
    {
        $errors = [];
        foreach ($this->files($source) as $path) {
            $layer = explode('/', substr($path, strlen($source) + 1))[0];
            if (!isset(self::ALLOWED[$layer])) {
                continue;
            }
            foreach (self::referencedNames((string) file_get_contents($path)) as $name) {
                if (!$this->allowed($layer, $name)) {
                    $errors[] = substr($path, strlen($source) + 1) . ": {$layer} must not depend on {$name}";
                }
            }
        }
        sort($errors);

        return array_values(array_unique($errors));
    }

    private function allowed(string $layer, string $name): bool
    {
        $parts = explode('\\', $name);
        if ($parts[0] === $this->root) {
            return in_array($parts[1] ?? '', self::ALLOWED[$layer], true);
        }

        return !in_array($layer, self::OWN_CODE_ONLY, true);
    }

    private function files(string $source): array
    {
        $files = [];
        $iterator = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($source, RecursiveDirectoryIterator::SKIP_DOTS));
        foreach ($iterator as $file) {
            if ($file instanceof SplFileInfo && $file->getExtension() === 'php') {
                $files[] = $file->getPathname();
            }
        }
        sort($files);

        return $files;
    }

    private static function referencedNames(string $code): array
    {
        $names = [];
        foreach (token_get_all($code) as $token) {
            if (is_array($token) && in_array($token[0], [T_NAME_QUALIFIED, T_NAME_FULLY_QUALIFIED], true)) {
                $names[] = ltrim($token[1], '\\');
            }
        }

        return $names;
    }
}
