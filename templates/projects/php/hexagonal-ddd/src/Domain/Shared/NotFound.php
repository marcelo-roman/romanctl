<?php

declare(strict_types=1);

namespace SampleApp\Domain\Shared;

use RuntimeException;

abstract class NotFound extends RuntimeException
{
}
