use std::sync::Arc;

use crate::adapters::http::Services;
use crate::adapters::logging::LogEventPublisher;
use crate::adapters::memory::orders::InMemoryOrderRepository;
use crate::application::clock::Clock;
use crate::application::orders::{EventPublisher, OrderService};
// romanctl:imports

pub fn in_memory(clock: Clock) -> Services {
    let events: Arc<dyn EventPublisher> = Arc::new(LogEventPublisher);
    Services {
        orders: Arc::new(OrderService::new(
            Arc::new(InMemoryOrderRepository::default()),
            events,
            clock,
        )),
        // romanctl:memory-services
    }
}
