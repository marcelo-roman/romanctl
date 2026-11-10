import { newIdentifier, parseIdentifier } from '../shared/identifier.js';

export class SampleItemId {
  private constructor(readonly value: string) {}

  static create(): SampleItemId {
    return new SampleItemId(newIdentifier());
  }

  static parse(value: string): SampleItemId {
    return new SampleItemId(parseIdentifier(value));
  }
}
