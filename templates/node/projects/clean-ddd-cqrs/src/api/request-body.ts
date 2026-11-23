import { BadRequestException } from '@nestjs/common';

export class RequestBody {
  private constructor(private readonly fields: Record<string, unknown>) {}

  static from(body: unknown): RequestBody {
    if (typeof body !== 'object' || body === null || Array.isArray(body)) {
      throw new BadRequestException('The request body must be a JSON object.');
    }
    return new RequestBody(body as Record<string, unknown>);
  }

  text(name: string): string {
    return this.optionalText(name) ?? '';
  }

  optionalText(name: string): string | undefined {
    const value = this.fields[name];
    if (value === undefined || value === null) {
      return undefined;
    }
    if (typeof value !== 'string') {
      throw new BadRequestException(`The field ${name} must be a string.`);
    }
    return value;
  }

  number(name: string): number {
    const value = this.fields[name];
    if (value === undefined || value === null) {
      return 0;
    }
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      throw new BadRequestException(`The field ${name} must be a number.`);
    }
    return value;
  }

  integer(name: string): number {
    const value = this.number(name);
    if (!Number.isInteger(value)) {
      throw new BadRequestException(`The field ${name} must be an integer.`);
    }
    return value;
  }
}
