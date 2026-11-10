use std::cmp::Reverse;
use std::collections::HashMap;

use async_trait::async_trait;
use tokio::sync::RwLock;

use crate::application::error::ApplicationError;
use crate::application::orders::OrderRepository;
use crate::domain::orders::Order;
use crate::domain::shared::Id;

#[derive(Default)]
pub struct InMemoryOrderRepository {
    orders: RwLock<HashMap<Id, Order>>,
}

#[async_trait]
impl OrderRepository for InMemoryOrderRepository {
    async fn save(&self, order: &Order) -> Result<(), ApplicationError> {
        self.orders.write().await.insert(order.id(), order.clone());
        Ok(())
    }

    async fn find(&self, id: Id) -> Result<Option<Order>, ApplicationError> {
        Ok(self.orders.read().await.get(&id).cloned())
    }

    async fn list(&self, limit: usize) -> Result<Vec<Order>, ApplicationError> {
        let mut orders: Vec<Order> = self.orders.read().await.values().cloned().collect();
        orders.sort_by_key(|order| Reverse(order.created_at()));
        orders.truncate(limit);
        Ok(orders)
    }
}
