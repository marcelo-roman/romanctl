use async_trait::async_trait;

use crate::application::error::ApplicationError;
use crate::domain::orders::{Order, OrderEvent};
use crate::domain::shared::Id;

#[async_trait]
pub trait OrderRepository: Send + Sync {
    async fn save(&self, order: &Order) -> Result<(), ApplicationError>;
    async fn find(&self, id: Id) -> Result<Option<Order>, ApplicationError>;
    async fn list(&self, limit: usize) -> Result<Vec<Order>, ApplicationError>;
}

pub trait EventPublisher: Send + Sync {
    fn publish(&self, events: &[OrderEvent]);
}
