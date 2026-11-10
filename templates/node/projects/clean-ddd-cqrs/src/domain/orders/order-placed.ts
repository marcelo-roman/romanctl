import type { DomainEvent } from '../shared/domain-event.js';

export class OrderPlaced implements DomainEvent {
  constructor(
    readonly orderId: string,
    readonly total: number,
    readonly currency: string,
    readonly occurredAt: Date,
  ) {}
}
