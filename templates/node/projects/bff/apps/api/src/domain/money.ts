import { DomainError } from './domain-error.js';

export const DEFAULT_CURRENCY = 'USD';

export class Money {
  private constructor(
    readonly cents: number,
    readonly currency: string,
  ) {}

  static fromCents(cents: number, currency: string = DEFAULT_CURRENCY): Money {
    if (!Number.isFinite(cents) || cents < 0) {
      throw new DomainError('The amount cannot be negative.');
    }
    const code = currency.trim().toUpperCase();
    if (!/^[A-Z]{3}$/.test(code)) {
      throw new DomainError('The currency must have 3 letters, like USD.');
    }
    return new Money(Math.round(cents), code);
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
      throw new DomainError('Cannot add different currencies.');
    }
    return new Money(this.cents + other.cents, this.currency);
  }
}
