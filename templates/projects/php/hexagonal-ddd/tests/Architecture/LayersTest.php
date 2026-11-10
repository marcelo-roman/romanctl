<?php

declare(strict_types=1);

namespace SampleApp\Tests\Architecture;

use PHPUnit\Framework\TestCase;

final class LayersTest extends TestCase
{
    private const string ROOT = 'SampleApp';

    private string $scratch = '';

    protected function tearDown(): void
    {
        if ($this->scratch !== '') {
            array_map(unlink(...), [$this->scratch . '/Domain/Model.php', $this->scratch . '/Application/Service.php']);
            array_map(rmdir(...), [$this->scratch . '/Domain', $this->scratch . '/Application', $this->scratch]);
        }
    }

    public function testDependenciesPointInward(): void
    {
        self::assertSame([], (new LayerRules(self::ROOT))->violations(dirname(__DIR__, 2) . '/src'));
    }

    public function testOutwardDependenciesAreReported(): void
    {
        $this->scratch = sys_get_temp_dir() . '/layers-' . bin2hex(random_bytes(4));
        mkdir($this->scratch . '/Domain', 0o777, true);
        mkdir($this->scratch . '/Application');
        file_put_contents($this->scratch . '/Domain/Model.php', "<?php\nuse Doctrine\\DBAL\\Connection;\nuse DateTimeImmutable;\n");
        file_put_contents($this->scratch . '/Application/Service.php', "<?php\nuse " . self::ROOT . "\\Adapters\\Http\\ApiResponse;\n");

        self::assertSame(
            [
                'Application/Service.php: Application must not depend on ' . self::ROOT . '\\Adapters\\Http\\ApiResponse',
                'Domain/Model.php: Domain must not depend on Doctrine\\DBAL\\Connection',
            ],
            (new LayerRules(self::ROOT))->violations($this->scratch),
        );
    }
}
