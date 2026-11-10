import { newIdentifier, parseIdentifier } from '../shared/identifier.js';

export class OrderId {
  private constructor(readonly value: string) {}

  static create(): OrderId {
    return new OrderId(newIdentifier());
  }

  static parse(value: string): OrderId {
    return new OrderId(parseIdentifier(value));
  }
}
