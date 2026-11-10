use std::sync::Arc;

use crate::adapters::http::Services;
use crate::adapters::logging::LogEventPublisher;
use crate::adapters::memory::orders::InMemoryOrderRepository;
use crate::adapters::postgres::Database;
use crate::adapters::postgres::orders::PostgresOrderRepository;
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

pub async fn postgres(url: &str, clock: Clock) -> Result<Services, sqlx::Error> {
    let database = Database::connect(url).await?;
    let events: Arc<dyn EventPublisher> = Arc::new(LogEventPublisher);
    Ok(Services {
        orders: Arc::new(OrderService::new(
            Arc::new(PostgresOrderRepository::new(database.clone())),
            events,
            clock,
        )),
        // romanctl:postgres-services
    })
}
