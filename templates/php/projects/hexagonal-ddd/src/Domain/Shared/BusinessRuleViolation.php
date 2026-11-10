<?php

declare(strict_types=1);

namespace SampleApp\Domain\Shared;

use DomainException;

final class BusinessRuleViolation extends DomainException
{
}
