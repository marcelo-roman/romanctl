import { DomainError } from './domain-error';

export const DEFAULT_CURRENCY = 'BRL';

export class Money {
  private constructor(
    readonly cents: number,
    readonly currency: string,
  ) {}

  static fromCents(cents: number, currency: string = DEFAULT_CURRENCY): Money {
    if (!Number.isFinite(cents) || cents < 0) {
      throw new DomainError('O valor não pode ser negativo.');
    }
    if (!/^[A-Za-z]{3}$/.test(currency.trim())) {
      throw new DomainError('A moeda deve ter 3 letras, como BRL.');
    }
    return new Money(Math.round(cents), currency.trim().toUpperCase());
  }

  static fromAmount(amount: number, currency: string = DEFAULT_CURRENCY): Money {
    return Money.fromCents(amount * 100, currency);
  }

  get amount(): number {
    return this.cents / 100;
  }

  multiply(quantity: number): Money {
    return new Money(this.cents * quantity, this.currency);
  }

  add(other: Money): Money {
    if (other.currency !== this.currency) {
      throw new DomainError('Não é possível somar moedas diferentes.');
    }
    return new Money(this.cents + other.cents, this.currency);
  }
}
