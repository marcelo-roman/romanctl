import type { DomainEvent } from './domain-event.js';

export abstract class AggregateRoot {
  private events: DomainEvent[] = [];

  protected raise(event: DomainEvent): void {
    this.events.push(event);
  }

  pullEvents(): DomainEvent[] {
    const events = this.events;
    this.events = [];
    return events;
  }
}
