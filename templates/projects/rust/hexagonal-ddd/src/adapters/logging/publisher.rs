use crate::application::orders::EventPublisher;
use crate::domain::orders::OrderEvent;

#[derive(Debug, Default)]
pub struct LogEventPublisher;

impl EventPublisher for LogEventPublisher {
    fn publish(&self, events: &[OrderEvent]) {
        for event in events {
            tracing::info!(name = event.name(), event = ?event, "domain event");
        }
    }
}
