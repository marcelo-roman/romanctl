import { BadRequestException } from '@nestjs/common';

export class RequestBody {
  private constructor(private readonly fields: Record<string, unknown>) {}

  static from(body: unknown): RequestBody {
    if (typeof body !== 'object' || body === null || Array.isArray(body)) {
      throw new BadRequestException('O corpo da requisição deve ser um objeto JSON.');
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
      throw new BadRequestException(`O campo ${name} deve ser um texto.`);
    }
    return value;
  }

  number(name: string): number {
    const value = this.fields[name];
    if (value === undefined || value === null) {
      return 0;
    }
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      throw new BadRequestException(`O campo ${name} deve ser um número.`);
    }
    return value;
  }

  integer(name: string): number {
    const value = this.number(name);
    if (!Number.isInteger(value)) {
      throw new BadRequestException(`O campo ${name} deve ser um número inteiro.`);
    }
    return value;
  }
}
